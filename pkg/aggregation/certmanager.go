package aggregation

import (
	"regexp"
	"strings"
)

func getCommonMessageForCertManager(reason string, message string) string {
	if strings.EqualFold(reason, "ErrGetKeyPair") {
		return "Error getting keypair for CA issuer"
	}
	return getCommonMessageForEvent(reason, message, issuerAggregationRegexps, issuerAggregationLabelValues)
}

var challengeAggregationRegexps = map[int]*regexp.Regexp{}
var challengeAggregationLabelValues = map[int]string{}

var challengeAggregations = map[string]string{
	"Error cleaning up challenge:.*":                     "Error cleaning up challenge",
	"Error presenting challenge:.*":                      "Error presenting challenge",
	"Presented challenge using .* challenge mechanism.*": "Presented challenge using acme challenge mechanism",
	"Domain .* verified with .* validation":              "Domain verified with validation",
	"Accepting challenge authorization failed:.*":        "Accepting challenge authorization failed",
}

var orderAggregationRegexps = map[int]*regexp.Regexp{}
var orderAggregationLabelValues = map[int]string{}

var orderAggregations = map[string]string{
	"Failed to determine a valid solver configuration for the set of domains on the Order:.*": "Failed to determine the list of Challenge resources needed for the Order",
	"Created Challenge resource .* for domain .*":                                             "Created Challenge resource for domain",
}
var certificateAggregationRegexps = map[int]*regexp.Regexp{}
var certificateAggregationLabelValues = map[int]string{}

var certificateAggregations = map[string]string{
	"The certificate request has failed to complete and will be retried.*":                                                                  "The certificate request has failed to complete and will be retried",
	"Regenerating private key due to change in fields: .*":                                                                                  "Regenerating private key due to change in fields",
	"Failed to decode private key stored in Secret .* - generating new key.*":                                                               "Failed to decode private key stored in Secret - generating new key",
	"User intervention required: existing private key in Secret .* does not match requirements on Certificate resource.*":                   "User intervention required: existing private key in Secret does not match requirements on Certificate resource",
	"Reusing private key stored in existing Secret resource.*":                                                                              "Reusing private key stored in existing Secret resource",
	"Stored new private key in temporary Secret resource.*":                                                                                 "Stored new private key in temporary Secret resource",
	"Failed to create CertificateRequest:.*":                                                                                                "Failed to create CertificateRequest",
	"Created new CertificateRequest resource.*":                                                                                             "Created new CertificateRequest resource",
	"Issuing certificate as Secret contains invalid private key data.*":                                                                     "Issuing certificate as Secret contains invalid private key data",
	"Issuing certificate as Secret contains an invalid certificate.*":                                                                       "Issuing certificate as Secret contains an invalid certificate",
	"Secret contains an invalid key-pair.*":                                                                                                 "Secret contains an invalid key-pair",
	"Existing private key is not up to date for spec.*":                                                                                     "Existing private key is not up to date for spec",
	"Issuing certificate as Secret was previously issued by.*":                                                                              "Issuing certificate as Secret was previously issued",
	"Secret was issued for .*. If this message is not transient, you might have two conflicting Certificates pointing to the same secret.*": "Secret was issued for another certificate",
}

var csrAggregationRegexps = map[int]*regexp.Regexp{}
var csrAggregationLabelValues = map[int]string{}

var csrAggregations = map[string]string{
	"Failed to decode CSR in spec.request:.*":                                                        "Failed to decode CSR in spec.request",
	"The CSR PEM requests a commonName that is not present in the list of dnsNames or ipAddresses.*": "The CSR PEM requests a commonName that is not present in the list of dnsNames or ipAddresses",
	"Failed to build order.*":                                         "Failed to build order",
	"Created Order resource.*":                                        "Created Order resource",
	"Failed to wait for order resource .* to become ready.*":          "Failed to wait for order resource to become ready",
	"Waiting on certificate issuance from order.*":                    "Waiting on certificate issuance from order",
	"Waiting for order-controller to add certificate data to Order.*": "Waiting for order-controller to add certificate data to Order resource",
	"Deleting Order with bad certificate.*":                           "Deleting Order with bad certificate",
	"Error updating certificate.*":                                    "Error updating certificate",
	"Referenced [Ss]ecret .* not found.*":                             "Referenced secret not found",
	"Failed to parse signing CA keypair from secret.*":                "Failed to parse signing CA keypair from secret",
	"Failed to get certificate key pair from secret.*":                "Failed to get certificate key pair from secret",
	"Error generating certificate template.*":                         "Error generating certificate template",
	"Error signing certificate.*":                                     "Error signing certificate",
	"Missing private key reference annotation.*":                      "Missing private key reference annotation",
	"Failed to parse signing key from secret.*":                       "Failed to parse signing key from secret",
	"Failed to get certificate CA key from secret.*":                  "Failed to get certificate CA key from secret",
	"Referenced.* not found":                                          "Referenced issuer not found",
	"Referenced.* is missing type":                                    "Referenced issuer is missing type",
	"Requester may not reference Namespaced Issuer.*":                 "Requester may not reference Namespaced Issuer",
	"Failed to parse requested duration.*":                            "Failed to parse requested duration",
	"CertificateSigningRequest minimum allowed duration is.*":         "CertificateSigningRequest duration is smaller than minimum allowed",
	"Failed to parse returned certificate bundle.*":                   "Failed to parse returned certificate bundle",
	"Referenced.* does not have a Ready status condition":             "Referenced issuer does not have a Ready status condition",
	"Failed to initialise vault client for signing.*":                 "Failed to initialise vault client for signing",
	"Vault failed to sign.*":                                          "Vault failed to sign",
	"Failed to initialise venafi client for signing.*":                "Failed to initialise venafi client for signing",
	"Failed to parse .* annotation.*":                                 "Failed to parse venafi annotation",
	"Failed to request venafi certificate.*":                          "Failed to request venafi certificate",
	"Failed to obtain venafi certificate.*":                           "Failed to obtain venafi certificate",
	"CSR .* has been approved":                                        "CSR has been approved",
}

var issuerAggregationRegexps = map[int]*regexp.Regexp{}
var issuerAggregationLabelValues = map[int]string{}

var issuerAggregations = map[string]string{
	"Failed to update ACME account.*":             "Failed to update ACME account",
	"Error initializing issuer.*":                 "Error initializing issuer",
	"Failed to parse existing ACME server URI.*":  "Failed to parse existing ACME server URI",
	"Failed to parse existing ACME account URI.*": "Failed to parse existing ACME account URI",
	"Failed to register ACME account.*":           "Failed to register ACME account",
}
