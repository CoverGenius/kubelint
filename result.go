package kubelint

import (
	log "github.com/sirupsen/logrus"
)

/**
*	Struct to carry all information necessary for the logger.
**/
type Result struct {
	Resource *YamlDerivedResource // the resource on which the rule was performed to get this result
	Message  string               // the complaining message (eg "no securityContextKey present")
	Level    log.Level            // the level of trouble this result causes
}
