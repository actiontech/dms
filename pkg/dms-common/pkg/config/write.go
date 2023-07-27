package config

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// WriteConfigFile saves the configuration representation to a file.
// The desired file permissions must be passed as in os.Open.
// The header is a string that is saved as a comment in the first line of the file.
func (c *ConfigFile) WriteConfigFile(fname string, perm uint32, header string, firstSections []string) (err error) {

	var file *os.File
	if file, err = os.Create(fname); err != nil {
		return err
	}
	defer file.Close()

	if err = c.Write(file, header, firstSections); err != nil {
		return err
	}
	file.Sync() //nolint:errcheck //we're OK if sync returns non-zero code
	return nil
}

// WriteConfigBytes returns the configuration file.
func (c *ConfigFile) WriteConfigBytes(header string) (config []byte) {
	buf := bytes.NewBuffer(nil)

	err := c.Write(buf, header, []string{})
	if err != nil && err != io.EOF {
		fmt.Println("WriteConfigBytes error: " + err.Error())
	}

	return buf.Bytes()
}

// Writes the configuration file to the io.Writer.
func (c *ConfigFile) Write(writer io.Writer, header string, firstSections []string) (err error) {
	buf := bytes.NewBuffer(nil)

	if header != "" {
		if _, err = buf.WriteString("# " + header + "\n"); err != nil {
			return err
		}
	}

	for _, section := range c.sectionOrder {
		sectionDataList, exist := c.data[section]
		if !exist {
			continue
		}
		if section == DefaultSection && len(sectionDataList) == 0 {
			continue // skip default section if empty
		}

		if section != virtualSection {
			if _, err = buf.WriteString("[" + section + "]\n"); err != nil {
				return err
			}
		}

		for _, kvData := range c.data[section] {
			switch kvData.key {
			case commentLine:
				if _, err := buf.WriteString(kvData.value + "\n"); err != nil {
					return err
				}
			case emptyLine:
				if _, err := buf.WriteString("\n"); err != nil {
					return err
				}
			default:
				if kvData.value == "" {
					if _, err := buf.WriteString(kvData.key + " = " + "\n"); err != nil {
						return err
					}
				} else {
					if _, err := buf.WriteString(kvData.key + " = " + kvData.value + "\n"); err != nil {
						return err
					}
				}
			}
		}
	}

	_, err = buf.WriteTo(writer)

	return err
}

// String return configuration as string.
func (c *ConfigFile) String() string {
	buf := bytes.NewBuffer(nil)
	for _, section := range c.sectionOrder {
		sectionmap, exist := c.data[section]
		if !exist {
			continue
		}
		if section == DefaultSection && len(sectionmap) == 0 {
			continue // skip default section if empty
		}
		if section != virtualSection {
			buf.WriteString("[" + section + "]\n")
		}

		for _, kvData := range c.data[section] {
			switch kvData.key {
			case commentLine:
				buf.WriteString(kvData.value + "\n")
			case emptyLine:
				buf.WriteString("\n")
			default:
				if kvData.value == "" {
					buf.WriteString(kvData.key + " = " + "\n")
				} else {
					buf.WriteString(kvData.key + " = " + kvData.value + "\n")
				}
			}
		}
	}
	return buf.String()
}
