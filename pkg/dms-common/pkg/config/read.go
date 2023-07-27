package config

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strings"
)

const BLANK_SECTION = "BLANK_SECTION"

// ReadConfigFile reads a file and returns a new configuration representation.
// This representation can be queried with GetString, etc.
// Not support multi-line value
func ReadConfigFile(fname string) (c *ConfigFile, err error) {
	var file *os.File

	if file, err = os.Open(fname); err != nil {
		return nil, err
	}

	c = NewConfigFile()
	if err = c.Read(file); err != nil {
		return nil, err
	}

	if err = file.Close(); err != nil {
		return nil, err
	}

	return c, nil
}

func ReadConfigBytes(conf []byte) (c *ConfigFile, err error) {
	buf := bytes.NewBuffer(conf)

	c = NewConfigFile()
	if err = c.Read(buf); err != nil {
		return nil, err
	}

	return c, err
}

// Read reads an io.Reader and returns a configuration representation. This
// representation can be queried with GetString, etc.
// Not support multi-line value
func (c *ConfigFile) Read(reader io.Reader) (err error) {
	buf := bufio.NewReader(reader)

	var (
		section, option     string
		getFirstRealSection bool
	)

	for {
		l, buferr := buf.ReadString('\n') // parse line-by-line
		l = strings.TrimSpace(l)

		if buferr != nil {
			if buferr != io.EOF {
				return err
			}

			if len(l) == 0 {
				break
			}
		}

		// switch written for readability (not performance)
		switch {
		case len(l) == 0: // empty line
			if getFirstRealSection {
				c.AddOption(section, emptyLine, "")
			}
			continue
		case l[0] == '#',
			l[0] == ';',
			len(l) >= 3 && strings.ToLower(l[0:3]) == "rem": // comment
			if getFirstRealSection {
				c.AddOption(section, commentLine, l, false)
			} else {
				c.AddOption(virtualSection, commentLine, l, false) //comments before first section
			}
			continue
		case l[0] == '[' && strings.Index(l, "]") > 0: // new section
			option = "" //nolint:ineffassign // 这段代码在for循环中将option变为空字符串，看起来是有意义的赋值，linter误报
			sectionEnd := strings.Index(l, "]")
			section = strings.TrimSpace(l[1:sectionEnd])
			if !getFirstRealSection {
				getFirstRealSection = true
			}
			c.AddSection(section)
		case section == "": // not new section and no section defined so far
			section = BLANK_SECTION

		default: // other alternatives
			i := strings.IndexAny(l, "=:")
			switch {
			case i > 0: // option and value
				i := strings.IndexAny(l, "=:")
				option = strings.TrimSpace(l[0:i])
				value := ""
				if i > 0 {
					value = strings.TrimSpace(StripComments(l[i+1:]))
				}
				c.AddOption(section, option, value, false)
			case i == -1:
				option = strings.TrimSpace(l)
				c.AddOption(section, option, "", false)
			default:
				return ReadError{CouldNotParse, l}
			}
		}

		// Reached end of file
		if buferr == io.EOF {
			break
		}
	}
	return nil
}

func StripComments(l string) string {
	// comments are preceded by space or TAB
	for _, c := range []string{" ;", "\t;", " #", "\t#"} {
		if i := strings.Index(l, c); i != -1 {
			l = l[0:i]
		}
	}
	return l
}
