package config

func (c *ConfigFile) SetString(section, option, value string) error {
	if section == "" {
		section = "default"
	}

	if _, ok := c.data[section]; ok {
		for _, kvData := range c.data[section] {
			if kvData.key == option {
				kvData.value = value
				return nil
			}
		}

		return GetError{OptionNotFound, "", "", section, option}
	}
	return GetError{SectionNotFound, "", "", section, option}
}
