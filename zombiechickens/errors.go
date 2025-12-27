package zombiechickens

import "fmt"

// StackValidationError wraps multiple stack validation errors
type StackValidationError struct {
	Errors []error
}

func (e *StackValidationError) Error() string {
	if len(e.Errors) == 0 {
		return "no stack validation errors"
	}
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}

	msg := fmt.Sprintf("%d stack validation errors:\n", len(e.Errors))
	for _, err := range e.Errors {
		msg += fmt.Sprintf("  - %s\n", err.Error())
	}
	return msg
}

// assertLegalStacks validates that all stacks in the farm follow legal stacking rules
func (f *Farm) assertLegalStacks() error {
	var errs []error

	for i, stack := range f.Stacks {
		// Check for empty stacks
		if len(stack) == 0 {
			errs = append(errs, fmt.Errorf("stack at index %d is empty", i))
			continue
		}

		// Count each item type in the stack
		counts := make(map[FarmItemType]int)
		for _, item := range stack {
			counts[item]++
		}

		// Validate the stack
		if err := validateStack(i, counts); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return &StackValidationError{Errors: errs}
	}
	return nil
}

// validateStack checks if a single stack follows legal stacking rules
func validateStack(index int, counts map[FarmItemType]int) error {
	// Determine what items are in the stack
	numTypes := len(counts)
	totalItems := 0
	for _, count := range counts {
		totalItems += count
	}

	// HayBale validation: must be 1-3 HayBales alone
	if hayCount, hasHay := counts[HayBale]; hasHay {
		if numTypes > 1 {
			return fmt.Errorf("stack at index %d contains HayBale with other item types", index)
		}
		if hayCount < 1 || hayCount > 3 {
			return fmt.Errorf("stack at index %d has %d HayBales but must have 1-3", index, hayCount)
		}
		return nil
	}

	// Scarecrow validation: must be exactly 1 Scarecrow alone
	if scarecrowCount, hasScarecrow := counts[Scarecrow]; hasScarecrow {
		if numTypes > 1 || scarecrowCount != 1 {
			return fmt.Errorf("stack at index %d must contain exactly 1 Scarecrow alone but has %d items", index, totalItems)
		}
		return nil
	}

	// BoobyTrap validation: must be exactly 1 BoobyTrap alone
	if boobyCount, hasBooby := counts[BoobyTrap]; hasBooby {
		if numTypes > 1 || boobyCount != 1 {
			return fmt.Errorf("stack at index %d must contain exactly 1 BoobyTrap alone but has %d items", index, totalItems)
		}
		return nil
	}

	// Shield validation: must be exactly 1 Shield alone
	if shieldCount, hasShield := counts[Shield]; hasShield {
		if numTypes > 1 || shieldCount != 1 {
			return fmt.Errorf("stack at index %d must contain exactly 1 Shield alone but has %d items", index, totalItems)
		}
		return nil
	}

	// WOLR validation: must be exactly 1 WOLR alone
	if wolrCount, hasWOLR := counts[WOLR]; hasWOLR {
		if numTypes > 1 || wolrCount != 1 {
			return fmt.Errorf("stack at index %d must contain exactly 1 WOLR alone but has %d items", index, totalItems)
		}
		return nil
	}

	// Shotgun/Ammo validation
	shotgunCount, hasShotgun := counts[Shotgun]
	_, hasAmmo := counts[Ammo]

	if hasShotgun && hasAmmo {
		// Must be exactly 1 Shotgun with any amount of Ammo, no other items
		if numTypes > 2 {
			return fmt.Errorf("stack at index %d contains Shotgun/Ammo with other illegal item types", index)
		}
		if shotgunCount != 1 {
			return fmt.Errorf("stack at index %d has %d Shotguns but must have exactly 1 with Ammo", index, shotgunCount)
		}
		return nil
	}

	if hasShotgun {
		// Shotgun alone (with 0 Ammo) is valid
		if numTypes > 1 {
			return fmt.Errorf("stack at index %d contains Shotgun with illegal item types (not Ammo)", index)
		}
		if shotgunCount != 1 {
			return fmt.Errorf("stack at index %d has %d Shotguns but must have exactly 1", index, shotgunCount)
		}
		return nil
	}

	if hasAmmo {
		// Ammo alone (any amount) is valid
		if numTypes > 1 {
			return fmt.Errorf("stack at index %d contains Ammo with illegal item types (not Shotgun)", index)
		}
		return nil
	}

	// Flamethrower/Fuel validation
	flamethrowerCount, hasFlamethrower := counts[Flamethrower]
	fuelCount, hasFuel := counts[Fuel]

	if hasFlamethrower && hasFuel {
		// Must be exactly 1 Flamethrower with exactly 1 Fuel, no other items
		if numTypes > 2 {
			return fmt.Errorf("stack at index %d contains Flamethrower/Fuel with other illegal item types", index)
		}
		if flamethrowerCount != 1 || fuelCount != 1 {
			return fmt.Errorf("stack at index %d must have exactly 1 Flamethrower and 1 Fuel but has %d Flamethrowers and %d Fuels", index, flamethrowerCount, fuelCount)
		}
		return nil
	}

	if hasFlamethrower {
		// Flamethrower alone (exactly 1) is valid
		if numTypes > 1 {
			return fmt.Errorf("stack at index %d contains Flamethrower with illegal item types (not Fuel)", index)
		}
		if flamethrowerCount != 1 {
			return fmt.Errorf("stack at index %d has %d Flamethrowers but must have exactly 1 when alone", index, flamethrowerCount)
		}
		return nil
	}

	if hasFuel {
		// Fuel alone (exactly 1) is valid
		if numTypes > 1 {
			return fmt.Errorf("stack at index %d contains Fuel with illegal item types (not Flamethrower)", index)
		}
		if fuelCount != 1 {
			return fmt.Errorf("stack at index %d has %d Fuels but must have exactly 1 when alone", index, fuelCount)
		}
		return nil
	}

	// If we get here, there's an unknown combination
	return fmt.Errorf("stack at index %d has unknown or invalid item combination", index)
}
