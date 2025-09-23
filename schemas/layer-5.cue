package schemas

import "time"

// #EnforcementAction is the central, auditable record of policy enforcement.
#EnforcementAction: {
	id: string
	timestamp: #Datetime
	target:                 #Target
	decision:               #Decision
	finding:                #Finding
	"remediation-plan-id"?: string @go(RemediationPlanId)
}

// #Decision is the high-level enforcement outcome.
#Decision: "Block" | "Mutate" | "Audit"

// #Finding is a self-contained record of a detected issue.
#Finding: {
	"requirement-id": string @go(RequirementId)
	result:           #Result
	message:          string
}

// #Target defines the subject of the enforcement action.
#Target: {
	"target-name": string @go(TargetName)
	"target-type": string @go(TargetType)
	"target-id"?:  string @go(TargetId)
}

// #Result is the outcome the assessment log.
#Result: "Not Run" | "Passed" | "Failed" | "Needs Review" | "Not Applicable" | "Unknown"

#Datetime: time.Format("2006-01-02T15:04:05Z07:00") @go(Datetime,format="date-time")

// Adapted from `darn/darnit` - https://github.com/kusari-oss/darn/blob/main/internal/core/models/models.go
// SPDX-License-Identifier: Apache-2.0

#RemediationPlan: {
	"id":             string
	"target":         string @go(Target)
	"enforcement-id": string @go(EnforcementId)
	"steps": [...#RemediationStep] @go(Steps)
}

#RemediationStep: {
	id:   string
	name: string
	parameters?: [string]: _
	reason:  string
	status?: "pending" | "running" | "success" | "failure"
	error?:  string
}
