package config

import (
	"strconv"
	"strings"
)

const (
	SectionDefault = "default"
)

// GetSections returns the list of sections in the configuration.
// (The default section always exists.)
func (c *ConfigFile) GetSections() (sections []string) {
	sections = make([]string, 0, len(c.data))
	for s := range c.data {
		sections = append(sections, s)
	}
	return sections
}

// HasSection checks if the configuration has the given section.
// (The default section always exists.)
func (c *ConfigFile) HasSection(section string) bool {
	if section == "" {
		section = SectionDefault
	}
	_, ok := c.data[strings.ToLower(section)]

	return ok
}

// GetOptions returns the list of options available in the given section.
// It returns an error if the section does not exist and an empty list if the section is empty.
// Options within the default section are also included.
func (c *ConfigFile) GetOptions(section string) (options []string, err error) {
	if section == "" {
		section = SectionDefault
	}
	section = strings.ToLower(section)

	if _, ok := c.data[section]; !ok {
		return nil, GetError{SectionNotFound, "", "", section, ""}
	}

	options = make([]string, 0, len(c.data[DefaultSection])+len(c.data[section]))
	for _, kvData := range c.data[DefaultSection] { //NO_DUPL_CHECK
		if kvData.key != emptyLine && kvData.key != commentLine {
			options = append(options, kvData.key)
		}
	}
	if section == SectionDefault {
		return options, nil
	}
	for _, kvData := range c.data[section] {
		if kvData.key != emptyLine && kvData.key != commentLine {
			options = append(options, kvData.key)
		}
	}

	return options, nil
}

// GetOptionsWithoutDefaultSection returns the list of options available in the given section.
// It returns an error if the section does not exist and an empty list if the section id empty.
// Options within the default section are not included.
func (c *ConfigFile) GetOptionsWithoutDefaultSection(section string) (options []string, err error) {
	section = strings.ToLower(section)

	if _, ok := c.data[section]; !ok {
		return nil, GetError{SectionNotFound, "", "", section, ""}
	}

	for _, kvData := range c.data[section] {
		options = append(options, kvData.key)
	}

	return options, nil
}

// HasOption checks if the configuration has the given option in the section.
// It returns false if either the option or section do not exist.
func (c *ConfigFile) HasOption(section string, option string) bool {
	if section == "" {
		section = SectionDefault
	}
	section = strings.ToLower(section)
	option = strings.ToLower(option)

	if _, ok := c.data[section]; !ok {
		return false
	}

	var okd, oknd bool
	for _, kvData := range c.data[DefaultSection] {
		if kvData.key == option {
			okd = true
		}
	}
	for _, kvData := range c.data[section] {
		if kvData.key == option {
			oknd = true
		}
	}

	return okd || oknd
}

// GetRawString gets the (raw) string value for the given option in the section.
// The raw string value is not subjected to unfolding, which was illustrated in the beginning of this documentation.
// It returns an error if either the section or the option do not exist.
// if has duplicated key, return the first one
func (c *ConfigFile) GetRawString(section string, option string) (value string, err error) {
	if section == "" {
		section = SectionDefault
	}

	section = strings.ToLower(section)
	option = strings.ToLower(option)
	if _, ok := c.data[section]; ok {

		for _, kvData := range c.data[section] {
			if kvData.key == option {
				return kvData.value, nil
			}
		}
		return "", GetError{OptionNotFound, "", "", section, option}
	}
	return "", GetError{SectionNotFound, "", "", section, option}
}

// GetString gets the string value for the given option in the section.
// If the value needs to be unfolded (see e.g. %(host)s example in the beginning of this documentation),
// then GetString does this unfolding automatically, up to DepthValues number of iterations.
// It returns an error if either the section or the option do not exist, or the unfolding cycled.
func (c *ConfigFile) GetString(section string, option string) (value string, err error) {
	value, err = c.GetRawString(section, option)
	if err != nil {
		return "", err
	}

	//trim ", ' left
	if strings.HasPrefix(value, "\"") {
		value = strings.TrimPrefix(value, "\"")
	} else {
		value = strings.TrimPrefix(value, "'")
	}

	//trim ", ' right
	value = strings.TrimSuffix(value, "\"")
	value = strings.TrimSuffix(value, "'")
	if strings.HasSuffix(value, "\"") {
		value = strings.TrimSuffix(value, "\"")
	} else {
		value = strings.TrimSuffix(value, "'")
	}

	return value, nil
}

// GetInt has the same behaviour as GetString but converts the response to int.
func (c *ConfigFile) GetInt(section string, option string) (value int, err error) {
	sv, err := c.GetString(section, option)
	if err == nil {
		value, err = strconv.Atoi(sv)
		if err != nil {
			err = GetError{CouldNotParse, "int", sv, section, option}
		}
	}

	return value, err
}

// GetFloat has the same behaviour as GetString but converts the response to float.
func (c *ConfigFile) GetFloat64(section string, option string) (value float64, err error) {
	sv, err := c.GetString(section, option)
	if err == nil {
		value, err = strconv.ParseFloat(sv, 64)
		if err != nil {
			err = GetError{CouldNotParse, "float64", sv, section, option}
		}
	}

	return value, err
}

// GetBool has the same behaviour as GetString but converts the response to bool.
// See constant BoolStrings for string values converted to bool.
func (c *ConfigFile) GetBool(section string, option string) (value bool, err error) {
	sv, err := c.GetString(section, option)
	if err != nil {
		return false, err
	}

	value, ok := BoolStrings[strings.ToLower(sv)]
	if !ok {
		return false, GetError{CouldNotParse, "bool", sv, section, option}
	}

	return value, nil
}

// get all key value data, if no existed section return false
func (c *ConfigFile) GetKVDatas(section string) ([]map[string]string, bool) {
	_, ok := c.data[section]
	if !ok {
		return nil, false
	}
	kvDatasMapList := make([]map[string]string, 0, len(c.data[section]))
	for _, kvdata := range c.data[section] {
		if kvdata.key == emptyLine || kvdata.key == commentLine {
			continue
		}
		kvDatasMapList = append(
			kvDatasMapList,
			map[string]string{kvdata.key: kvdata.value},
		)
	}
	return kvDatasMapList, true
}
