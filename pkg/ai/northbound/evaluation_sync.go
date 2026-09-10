package northbound

import (
	"fmt"
	"net/url"
	"strings"
	"unicode"
)

func validateEvaluationModelAssociation(targetURI, resultURI string) error {
	for _, reference := range []struct{ source, uri string }{
		{"target model", targetURI},
		{"evaluation result", resultURI},
	} {
		parsed, err := url.Parse(reference.uri)
		if err != nil || reference.uri == "" || strings.IndexFunc(reference.uri, unicode.IsSpace) >= 0 ||
			!parsed.IsAbs() || (parsed.Host == "" && parsed.Path == "" && parsed.Opaque == "") {
			return fmt.Errorf("%s modelURI must be a valid nonempty absolute URI", reference.source)
		}
	}
	// Storage references are identifiers: normalization could conflate distinct models.
	if targetURI != resultURI {
		return fmt.Errorf("evaluation result modelURI does not match the target model")
	}
	return nil
}
