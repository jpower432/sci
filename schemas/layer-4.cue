package schemas

#Layer4: {
    evaluations: [#ControlEvaluation, ...#ControlEvaluation]
}

// Types

#ControlEvaluation: {
    name: string
    // frameworkID contains the unique identifier of the framework being evaluated
    // TODO: in the case of versioned frameworks (ex: NIST 800-53), should we expect the version to be part of the ID?
    frameworkID: string
    // one or more evaluations of the framework controls
    evaluations: [#ControlEvaluation, ...#ControlEvaluation]
}

// URL describes a specific subset of URLs of interest to the framework
// TODO: this definition should be imported from a more appropriate module/package
#URL: =~"^https?://[^\\s]+$"

// ControlEvaluation describes the evaluation of the layer 2 control referenced by controlID and the assessment of that control's requirements.
#ControlEvaluation: {
    // ID of the layer 2 control being evaluated
    controlID: string
    // TODO: should there also be a frameworkID here to make a ControlEvaluation more self-contained?
    // one or more assessments of the control requirements
    // TODO: should it be 0 or more to account for an evaluation where planning is "in-progress" for which assessments should be run?
    assessments: [#Assessment, ...#Assessment]
}

// Assessment describes the evaluation of layer 2 control requirement referenced by requirementID and the assessment methods used to assess that requirement.
#Assessment: {
    // TODO: should there also be frameworkID and controlID here to make a Assessment more self-contained?
    // ID of the layer 2 control requirement being evaluated
    requirementID: string
    // the methods used to assess the requirement
    methods: [#AssessmentMethod, ...#AssessmentMethod]
}

// AssessmentMethod describes the method used to assess the layer 2 control requirement referenced by requirementID.
#AssessmentMethod: {
    // TODO: should there also be frameworkID and controlID here to make a AssessmentMethod more self-contained?
    // Name is the name of the method used to assess the requirement.
    name: string
    // Description is a detailed explanation of the method.
    description: string
    // Run is a boolean indicating whether the method was run or not. When run is true, result is expected to be present.
    run: bool
    // Remediation guide is a URL to remediation guidance associated with the control's assessment requirement and this specific assessment method.
    remediation_guide?: #URL
}

// See https://cuelang.org/docs/tour/types/sumstruct/
#AssessmentMethod: {
    name: string
    description: string
    run: false
    remediation_guide?: #URL
} | {
    name: string
    description: string
    run: true
    remediation_guide?: #URL
    result!: #AssessmentResult
}

// AssessmentResult describes the result of the assessment of a layer 2 control requirement.
#AssessmentResult: {
    // status describes what happened when the assessment method was run
    //  * passed when all evidence suggests the control is met
    //  * failed when some evidence suggests the control is not met
    //  * needs_review when evidence was gathered but cannot be reliably interpreted to reach a decision. A human should review the evidence gathered
    //  * error when the method failed to execute
    status: #Status
    // TODO: I can imagine assessment methods potentially making more than a single change, perhaps this should be a list
    change?: #Change
}

// Status constrains the acceptable values describing the result of the assessment of a level 2 control requirement.
#Status: "passed" | "failed" | "needs_review" | "error"

// Change describes whether the execution of an automated assessment of a layer 2 control requirement resulted in changes being made to the system(s) under assessment.
// TODO: flesh out more once we have one or more examples of existing usage/dependency/necessity
#Change: {
    // TODO: document all fields here with more clarity once we have one or more examples of existing usage/dependency/necessity
    // target name is ¯\_(ツ)_/¯
    "target-name": string
    // applied describes whether the change was applied to the system(s) under assessment
    applied: bool
    // reverted describes whether the change was reverted from the system(s) under assessment
    reverted: bool
    // error describes whether an error occurred during either the application or reversion of the change
    error?: string
    // target object is ¯\_(ツ)_/¯
    "target-object"?: _
}
