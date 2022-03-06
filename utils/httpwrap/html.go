package httpwrap

import ()

var templateDir *string

// SetTemplateDir ...
func SetTemplateDir(dir string) { *templateDir = dir }

// GetTemplateDir ...
func GetTemplateDir(dir string) string { return *templateDir }
