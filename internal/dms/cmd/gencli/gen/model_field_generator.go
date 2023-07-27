package gen

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mindstand/gogm/v2/cmd/gogmcli/util"
)

type dataNodeTplRelConf struct {
	StructName      string
	StructFieldName string

	// 属性的gorm名称，如` gorm:"column:xxx; not null"`中的xxx
	StructFieldGormName string
}

func parseModel(debug bool, directory string) (nodeFieldFuncs map[string][]*dataNodeTplRelConf, imports []string, packageName string, err error) {
	fieldConfs := make(map[string][]*fieldConf, 0)
	imps := map[string][]string{}
	packageName = ""

	err = filepath.Walk(directory, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info == nil {
			return errors.New("file info is nil")
		}

		if info.IsDir() && filePath != directory {
			if debug {
				log.Printf("skipping [%s] as it is a directory\n", filePath)
			}
			return filepath.SkipDir
		}

		if path.Ext(filePath) == ".go" {
			if debug {
				log.Printf("parsing go file [%s]\n", filePath)
			}
			err := parseFile(filePath, imps, &packageName, &fieldConfs, debug)
			if err != nil {
				if debug {
					log.Printf("failed to parse go file [%s] with error '%s'\n", filePath, err.Error())
				}
				return err
			}
			if debug {
				log.Printf("successfully parsed go file [%s]\n", filePath)
			}
		} else if debug {
			log.Printf("skipping non go file [%s]\n", filePath)
		}

		return nil
	})
	if err != nil {
		return nil, nil, "", err
	}

	for _, imp := range imps {
		imports = append(imports, imp...)
	}

	imports = util.RemoveDuplicates(imports)

	for i := 0; i < len(imports); i++ {
		imports[i] = strings.Replace(imports[i], "\"", "", -1)
	}

	nodeFieldFuncs = make(map[string][]*dataNodeTplRelConf)

	for nodeName, confs := range fieldConfs {

		nodeFieldFuncs[nodeName] = make([]*dataNodeTplRelConf, len(confs))
		for i, conf := range confs {
			nodeFieldFuncs[nodeName][i] = &dataNodeTplRelConf{
				StructName:          nodeName,
				StructFieldName:     conf.StructFieldName,
				StructFieldGormName: conf.GormFieldName,
			}
		}
		if debug {
			log.Printf("adding node to nodeFieldFuncs node [%s]", nodeName)
		}
	}

	return nodeFieldFuncs, imports, packageName, nil
}

// parses each file using ast looking for nodes to handle
func parseFile(filePath string, imports map[string][]string, packageName *string, fieldConfs *map[string][]*fieldConf, debug bool) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	if node.Scope != nil {
		*packageName = node.Name.Name
		if node.Scope.Objects != nil && len(node.Scope.Objects) != 0 {
			for label, config := range node.Scope.Objects {
				log.Println("checking ", label)
				tSpec, ok := config.Decl.(*ast.TypeSpec)
				if !ok {
					continue
				}

				strType, ok := tSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}

				if node.Imports != nil && len(node.Imports) != 0 {
					var imps []string

					for _, impSpec := range node.Imports {
						imps = append(imps, impSpec.Path.Value)
					}

					imports[label] = imps
				}

				if debug {
					log.Printf("node [%s] strType [%v]", label, strType)
				}

				err = parseNode(strType, fieldConfs, label, fset)
				if err != nil {
					return err
				}

			}
		}
	}

	return nil
}

type fieldConf struct {
	StructFieldName string
	GormFieldName   string
}

var gormRex = regexp.MustCompile(`gorm:"(.*?)"`)

// parseNode generates configuration for struct fields
func parseNode(strType *ast.StructType, fieldConfs *map[string][]*fieldConf, label string, fset *token.FileSet) error {
	if strType.Fields != nil && strType.Fields.List != nil && len(strType.Fields.List) != 0 {
	fieldLoop:
		for _, field := range strType.Fields.List {
			if len(field.Names) == 0 && fmt.Sprintf("%v", field.Type) == "Model" {
				(*fieldConfs)[label] = append((*fieldConfs)[label], &fieldConf{
					StructFieldName: "UID",
					GormFieldName:   "uid",
				})
			}
			if field.Tag != nil && field.Tag.Value != "" {
				gromPart := gormRex.FindStringSubmatch(field.Tag.Value)

				if len(gromPart) == 0 {
					continue fieldLoop
				}

				// if !strings.Contains(gromPart[1], "not null") {
				// 	continue fieldLoop
				// }

				structFieldName := field.Names[0].Name
				gormFieldName := strings.ToLower(structFieldName)
				if strings.Contains(gromPart[1], "column:") {
					parts := strings.Split(gromPart[1], ";")
					for _, p := range parts {
						if strings.Contains(p, "column:") {
							gormFieldName = util.RemoveFromString(p, "column:", "\"")
						}
					}
				}

				(*fieldConfs)[label] = append((*fieldConfs)[label], &fieldConf{
					StructFieldName: field.Names[0].Name,
					GormFieldName:   gormFieldName,
				})
			}
		}
	}

	return nil
}
