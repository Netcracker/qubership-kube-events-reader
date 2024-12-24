package utils

import (
	"fmt"
	"regexp"
	"strings"
)

type NamespaceFlagsType []string

func (i *NamespaceFlagsType) String() string {
	return strings.Join(*i, ",")
}

var namespaceValidator = regexp.MustCompile("^[a-z0-9]([-a-z0-9]*[a-z0-9])?$")

func (i *NamespaceFlagsType) Set(value string) error {
	if !namespaceValidator.MatchString(value) {
		return fmt.Errorf("namespace is not valid. Got string: %s", value)
	}
	for _, ns := range *i {
		if strings.Compare(ns, value) == 0 {
			return nil
		}
	}
	*i = append(*i, value)
	return nil
}

type SinksFlagsType []string

func (i *SinksFlagsType) String() string {
	return strings.Join(*i, ",")
}

var outputsValidator = regexp.MustCompile("^metrics|logs$")

func (i *SinksFlagsType) Set(value string) error {
	if !outputsValidator.MatchString(value) {
		return fmt.Errorf("output value is not valid. Got string: %s", value)
	}
	for _, sink := range *i {
		if strings.Compare(sink, value) == 0 {
			return nil
		}
	}
	*i = append(*i, value)
	return nil
}
