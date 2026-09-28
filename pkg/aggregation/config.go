package aggregation

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// DefaultMaxMessageLength is the maximum number of characters in the message label
// when Config.MaxMessageLength is not set.
const DefaultMaxMessageLength = 256

// combinedEventsPrefix is added by the client-go event correlator when it merges similar events.
const combinedEventsPrefix = "(combined from similar events): "

const truncationSuffix = "..."

// Config holds user-defined message aggregation settings.
type Config struct {
	// MaxMessageLength is the maximum number of characters in the message label. Zero means DefaultMaxMessageLength.
	MaxMessageLength int `json:"maxMessageLength,omitempty"`
	// Rules collapse messages into a fixed label value. They are checked in order, before the built-in patterns.
	Rules []Rule `json:"rules,omitempty"`
}

// Rule maps event messages that match all set regular expressions to Value.
type Rule struct {
	Kind    string `json:"kind,omitempty"`
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message"`
	// Value is the label value. It can reference capture groups of Message, for example ${1}.
	Value string `json:"value"`
}

type compiledRule struct {
	kind    *regexp.Regexp
	reason  *regexp.Regexp
	message *regexp.Regexp
	value   string
}

var (
	customRules      []compiledRule
	maxMessageLength = DefaultMaxMessageLength
)

func applyConfig(config *Config) error {
	customRules = nil
	maxMessageLength = DefaultMaxMessageLength
	if config == nil {
		return nil
	}
	if config.MaxMessageLength < 0 {
		return fmt.Errorf("maxMessageLength must not be negative, got %d", config.MaxMessageLength)
	}
	if config.MaxMessageLength > 0 {
		maxMessageLength = max(config.MaxMessageLength, len(truncationSuffix)+1)
	}
	rules := make([]compiledRule, 0, len(config.Rules))
	var errs []error
	for i, rule := range config.Rules {
		compiled, err := compileRule(rule)
		if err != nil {
			errs = append(errs, fmt.Errorf("rule %d: %w", i, err))
			continue
		}
		rules = append(rules, compiled)
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	customRules = rules
	return nil
}

func compileRule(rule Rule) (compiledRule, error) {
	if rule.Message == "" {
		return compiledRule{}, errors.New("message is required")
	}
	if rule.Value == "" {
		return compiledRule{}, errors.New("value is required")
	}
	compiled := compiledRule{value: rule.Value}
	var err error
	if compiled.message, err = regexp.Compile(rule.Message); err != nil {
		return compiledRule{}, fmt.Errorf("invalid message regexp: %w", err)
	}
	if rule.Kind != "" {
		if compiled.kind, err = regexp.Compile(rule.Kind); err != nil {
			return compiledRule{}, fmt.Errorf("invalid kind regexp: %w", err)
		}
	}
	if rule.Reason != "" {
		if compiled.reason, err = regexp.Compile(rule.Reason); err != nil {
			return compiledRule{}, fmt.Errorf("invalid reason regexp: %w", err)
		}
	}
	return compiled, nil
}

func getMessageByCustomRules(kind string, reason string, message string) (string, bool) {
	for _, rule := range customRules {
		if rule.kind != nil && !rule.kind.MatchString(kind) {
			continue
		}
		if rule.reason != nil && !rule.reason.MatchString(reason) {
			continue
		}
		match := rule.message.FindStringSubmatchIndex(message)
		if match == nil {
			continue
		}
		return string(rule.message.ExpandString(nil, rule.value, message, match)), true
	}
	return "", false
}

// truncateMessage keeps the first line of the message and limits it to maxMessageLength characters.
// It cuts by runes, not bytes: the Prometheus client panics on label values that are not valid UTF-8.
func truncateMessage(message string) string {
	if i := strings.IndexAny(message, "\r\n"); i >= 0 {
		message = message[:i]
	}
	message = strings.TrimSpace(message)
	if utf8.RuneCountInString(message) <= maxMessageLength {
		return message
	}
	runes := []rune(message)
	return strings.TrimSpace(string(runes[:maxMessageLength-len(truncationSuffix)])) + truncationSuffix
}
