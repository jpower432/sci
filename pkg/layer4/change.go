package layer4

import (
	"fmt"
)

type ApplyFunc func() (*interface{}, error)
type RevertFunc func() error

// Change is a struct that contains the data and functions associated with a single change
type Change struct {
	Target_Name   string       // TargetName is the name or ID of the resource or configuration that is to be changed
	Target_Object *interface{} // TargetObject is the object that is to be changed, retained here for logging purposes
	Applied       bool         // Applied is true if the change was successfully applied at least once
	Reverted      bool         // Reverted is true if the change was successfully reverted and not applied again
	Error         error        // Error is used if any error occurred during the change

	applyFunc  ApplyFunc  // applyFunc is the function that will be executed to make the change
	revertFunc RevertFunc // revertFunc is the function that will be executed to undo the change
}

// NewChange creates a new Change struct with the provided data
func NewChange(targetName string, targetObject *interface{}, applyFunc ApplyFunc, revertFunc RevertFunc) *Change {
	return &Change{
		Target_Name:   targetName,
		Target_Object: targetObject,
		applyFunc:     applyFunc,
		revertFunc:    revertFunc,
	}
}

// Apply executes the Apply function for the change
func (c *Change) Apply() {
	err := c.precheck()
	if err != nil {
		c.Error = err
		return
	}
	// Do nothing if the change has already been applied and not reverted
	if c.Applied && !c.Reverted {
		return
	}
	obj, err := c.applyFunc()
	if err != nil {
		c.Error = err
		return
	}
	if obj != nil {
		c.Target_Object = obj
	}
	c.Applied = true
	c.Reverted = false
}

// Revert executes the Revert function for the change
func (c *Change) Revert() {
	err := c.precheck()
	if err != nil {
		c.Error = err
		return
	}
	// Do nothing if the change has not been applied
	if !c.Applied {
		return
	}
	err = c.revertFunc()
	if err != nil {
		c.Error = err
		return
	}
	c.Reverted = true
}

// precheck verifies that the applyFunc and revertFunc are defined for the change
func (c *Change) precheck() error {
	if c.applyFunc == nil {
		return fmt.Errorf("no apply function defined for change")
	} else if c.Error != nil {
		return fmt.Errorf("change has a previous error and can no longer be applied")
	}
	if c.revertFunc == nil {
		return fmt.Errorf("no revert function defined for change")
	}
	return nil
}
