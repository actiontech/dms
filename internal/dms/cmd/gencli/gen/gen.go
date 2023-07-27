package gen

import (
	"bytes"
	"fmt"
	"go/format"
	"log"
	"os"
	"path"
	"sort"
	textTpl "text/template"
)

type dataNodeTemplateConfig struct {
	NodeFieldFuncs map[string][]*dataNodeTplRelConf
	NodeNames      []*dataNodeTplNodeName
}

type dataNodeTplNodeName struct {
	NodeName string
}

// GenRepoFieldsFile 根据model中的node定义，来生成biz层的repo_fields文件，包含所有node的属性field名称，是专为dms项目实现order by 和 filter by使用的
func GenRepoFieldsFile(debug bool, modelDir, targetDir string) error {
	funcs, _, _, err := parseModel(debug, modelDir)
	if nil != err {
		return err
	}

	//write templates out
	tpl := textTpl.New("fieldFile")

	//register templates
	for _, templateString := range []string{fieldFileTempl, fieldFileFieldSpec, fieldFileNodeSpec} {
		var err error
		tpl, err = tpl.Parse(templateString)
		if err != nil {
			return fmt.Errorf("parse fieldFileTempl template error: %v", err)
		}
	}

	if len(funcs) == 0 {
		log.Printf("no functions to write, exiting")
		return nil
	}

	buf := new(bytes.Buffer)
	err = tpl.Execute(buf, dataNodeTemplateConfig{
		NodeFieldFuncs: funcs,
		NodeNames:      genDataNodeName(funcs),
	})
	if err != nil {
		return err
	}

	// format generated code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}

	// create the file
	f, err := os.Create(path.Join(targetDir, "repo_fields.go"))
	if err != nil {
		return err
	}

	// write code to the file
	lenBytes, err := f.Write(formatted)
	if err != nil {
		return err
	}

	if debug {
		log.Printf("done after writing [%v] bytes!", lenBytes)
	}

	// close the buffer
	err = f.Close()
	if err != nil {
		return err
	}

	log.Printf("wrote repo functions to file [%s/repo_fields.go]", targetDir)

	return nil
}

func genDataNodeName(nodeFieldFuncs map[string][]*dataNodeTplRelConf) []*dataNodeTplNodeName {
	nodeNames := make([]*dataNodeTplNodeName, 0, len(nodeFieldFuncs))

	for nodeName := range nodeFieldFuncs {
		nodeNames = append(nodeNames, &dataNodeTplNodeName{
			NodeName: nodeName,
		})
	}

	sort.Slice(nodeNames, func(i, j int) bool {
		return nodeNames[i].NodeName < nodeNames[j].NodeName
	})
	return nodeNames
}
