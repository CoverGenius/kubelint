package kubelint

import (
	log "github.com/sirupsen/logrus"
)

//	Result carries all information necessary for the logger.
type Result struct {
	Resources []*YamlDerivedResource // the resource(s) on which the rule was performed to get this result
	Message   string                 // the complaining message (eg "no securityContextKey present")
	Level     log.Level              // the level of trouble this result causes
}
