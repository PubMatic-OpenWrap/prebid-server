package config

import (
	"errors"
	"fmt"
	"strings"

	validator "github.com/asaskevich/govalidator"
)

// createElements contains list of valid VAST events
var createElements = [...]string{"impression", "tracking", "clicktracking", "companionclickthrough", "error", "nonlinearclicktracking"}

// eventTypes contains list of valid VAST event types for tracking element
var eventTypes = [...]string{"start", "firstQuartile", "midPoint", "thirdQuartile", "complete"}

// Validate validates the events and returns error if at least one is invalid.
func (e Events) Validate() error {
	if e.Enabled { // validate only if events are enabled
		if !isValidURL(e.DefaultURL) {
			return errors.New("Invalid Events.DefaultURL")
		}
		err := validateVASTEvents(e.VASTEvents)
		if err != nil {
			return err
		}
	}
	return nil // valid events or events are not enabled skip validation
}

func validateVASTEvents(events []VASTEvent) error {
	if nil != events {
		for i, event := range events {
			if err := validateVASTEvent(event, i); err != nil {
				return err
			}
		}
	}
	return nil
}

// validateVASTEvent validates event object and  returns error if at least one is invalid
func validateVASTEvent(event VASTEvent, index int) error {
	if !isValidCreateElement(event.CreateElement) {
		return fmt.Errorf("Invalid VASTEvents[%d].CreateElement", index)
	}

	// VASTEvent.ExcludeDefaultURL assumed to be false by default
	if !isValidType(event) {
		if isTrackingEvent(event) {
			return fmt.Errorf("Missing or Invalid VASTEvents[%d].Type.Valid values are %v", index, strings.Join(eventTypes[:], " ,"))
		}
		return fmt.Errorf("VASTEvents[%d].Type is not applicable for create element '%s'", index, event.CreateElement)
	}
	for i, url := range event.URLs {
		if !isValidURL(url) {
			return fmt.Errorf("Invalid VASTEvents[%d].URLs[%d]", index, i)
		}
	}
	// ensure at least one valid url exists when default URL to be excluded
	if event.ExcludeDefaultURL && len(event.URLs) == 0 {
		return fmt.Errorf("Please provide at least one valid URL for VASTEvents[%d] or set VASTEvents[%d].ExcludeDefaultURL=false", index, index)
	}

	return nil // no errors
}

// isValidCreateElement checks if value belongs to
//  createElements
func isValidCreateElement(element string) bool {
	valid := false
	// validate create element
	for _, validEle := range createElements {
		if element == validEle {
			valid = true
			break
		}
	}
	return valid
}

// isValidtype checks valid type is provided in case event is of type tracking
// in case of other events this value must be empty
func isValidType(event VASTEvent) bool {
	if isTrackingEvent(event) {
		for _, validType := range eventTypes {
			if event.Type == validType {
				// valid event type for create element tracking
				return true
			}
		}
		return false // invalid event type for create element tracking
	}
	return len(event.Type) == 0 // event.type must be empty in case create element is not tracking
}

// isValidURL validates the event URL
func isValidURL(eventURL string) bool {
	return validator.IsURL(eventURL) && validator.IsRequestURL(eventURL)
}

// isTrackingEvent returns true if event object contains event.CreateElement == "tracking"
func isTrackingEvent(event VASTEvent) bool {
	return event.CreateElement == "tracking"
}
