package kubelint

import "fmt"

// This object is used to store all the rules belonging to a resource group and looks like:

//rulesorter.ruleSorter{
//rules:24:(*lint.Rule)(0xc00039caf0),
//edges:24:map[lint.RuleID]lint.RuleID{}
//
type ruleSorter struct {
	rules map[RuleID]*rule
	edges map[RuleID]map[RuleID]RuleID
}

// Retrieve the rule given its ID
// May as well implement this since I have to make a map for other operations anyway
func (r *ruleSorter) get(id RuleID) *rule {
	return r.rules[id]
}

func (r *ruleSorter) clone() *ruleSorter {
	edgesClone := make(map[RuleID]map[RuleID]RuleID)
	rulesClone := make(map[RuleID]*rule)

	for id, rule := range r.rules {
		rulesClone[id] = rule
	}
	for id, predecessors := range r.edges {
		edgesClone[id] = make(map[RuleID]RuleID)
		for incoming, _ := range predecessors {
			edgesClone[id][incoming] = incoming
		}
	}
	return &ruleSorter{edges: edgesClone, rules: rulesClone}
}

// Create a new ruleSorter given a list of rules
// Usual use case is to use the ruleSorter to access the rules in the correct order!
func newRuleSorter(rules []*rule) *ruleSorter {
	e := make(map[RuleID]map[RuleID]RuleID)
	r := make(map[RuleID]*rule)
	for _, rule := range rules {
		r[rule.ID] = rule
		e[rule.ID] = make(map[RuleID]RuleID)
		for _, prereq := range rule.Prereqs {
			e[rule.ID][prereq] = prereq
		}
	}
	return &ruleSorter{edges: e, rules: r}
}

func (r *ruleSorter) getDependentRules(masterId RuleID) []*rule {
	ruleIDs := r.getDependents(masterId)
	var rules []*rule
	for _, id := range ruleIDs {
		rules = append(rules, r.rules[id])
	}
	return rules
}

//	Given a rule (identified by its ID), get all the rules that are dependent upon it.
//   This implies that those rules' Condition functions are keeping a reference to the same struct.
// 	Ie, you would never have a rule dependent on another if they are referring to different objects.
func (r *ruleSorter) getDependents(masterId RuleID) []RuleID {
	var dependentIDs []RuleID
	for id := range r.rules {
		for _, masterRuleID := range r.rules[id].Prereqs {
			if masterRuleID == masterId {
				dependentIDs = append(dependentIDs, id)
				dependentIDs = append(dependentIDs, r.getDependents(id)...)
			}
		}
	}
	return dependentIDs
}

// Use this when you want to retrieve AND get rid of all rules that are dependent on a particular rule.
// Usually you want to use this when a rule fails, and you would like to avoid executing
// the rules that depend on this rule's success.
func (r *ruleSorter) popDependentRules(masterId RuleID) []*rule {
	dependents := r.getDependentRules(masterId)
	// now just delete them from the map.
	for _, rule := range dependents {
		delete(r.edges, rule.ID)
	}
	return dependents
}

func (r *ruleSorter) isEmpty() bool {
	return len(r.edges) == 0
}

//	This method removes the given rule from the ruleSorter structure.
//	For example, when a rule is satisfied and we don't have to worry about the fix
//	methods attached to a rule, remove the rule from the structure.
//	Anyone dependent upon this rule will be fine, since the rule is satisfied. So
//	they can all safely execute their fixes.
//	The rule is removed from the edges map and all rules depending on this one have it removed from their edges.
func (r *ruleSorter) remove(id RuleID) {
	delete(r.edges, id)
	// it's still maintained in the rule map and that's fine!
	for _, dependentId := range r.getDependents(id) {
		delete(r.edges[dependentId], id)
	}
}

// When you need to know which rule you should execute next, call this method. It will remove
// the rule from the data structure and return it.
// The algorithm is as follows:
//
//1. Find a rule with no dependencies, in case of multiple such rules the first one is chosen
//2. Find all the rules which depend on this rule, and remove it from it's dependency list
//3. remove the rule itself from the edge map
//4. Return the rule
func (r *ruleSorter) popNextAvailable() *rule {
	var ruleId RuleID
	cycle := true
	for id, incoming := range r.edges {
		if len(incoming) == 0 {
			ruleId = id
			cycle = false
			break
		}
	}
	// If we don't have any empty edges list, that means
	// we have a cycle somewhere
	if cycle {
		for id, edges := range r.edges {
			fmt.Printf("%s:\n", id)
			for rule, _ := range edges {
				fmt.Printf("\t%s\n", rule)
			}
		}
		panic("Either there's a cycle in your dependencies OR you've forgotten to include a prerequisite rule. Please be more careful")
	}
	for _, id := range r.getDependents(ruleId) {
		// update their edges so that they don't remember ruleId anymore!
		delete(r.edges[id], ruleId)
	}
	// now please forget totally about this ruleID from the edges
	delete(r.edges, ruleId)
	// its map is also gone, (it would have been empty anyways)
	return r.rules[ruleId]
}
