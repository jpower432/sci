# OSPS Checklist Example

This directory contains an example demonstrating how to generate an OSPS (Open Source Project Security) Baseline checklist from a Layer 3 policy.

## Files

- `osps-example-policy.yaml` - Example Layer 3 policy that references OSPS-B controls
- `osps_checklist_example.go` - Go program that loads a policy and generates a markdown checklist
- `osps-checklist-example.md` - Generated example checklist output

## Usage

To generate a checklist from a policy file:

```bash
cd examples
go run osps_checklist_example.go osps-example-policy.yaml
```

Or specify a different policy file:

```bash
go run osps_checklist_example.go path/to/your-policy.yaml
```

## How It Works

1. **Loads the Layer 3 Policy**: The program loads a YAML policy file that references OSPS-B controls
2. **Resolves Catalog References**: The policy's `mapping-references` section contains a URL pointing to the OSPS-B catalog (Layer 2)
3. **Loads Layer 2 Catalog**: The system loads the OSPS-B catalog from the specified URL
4. **Applies Policy Modifications**: For each control referenced in the policy:
   - Finds the control in the loaded catalog
   - Applies any modifications specified in the policy (title overrides, etc.)
   - Creates checklist items from the control's assessment requirements
   - Applies assessment requirement modifications if specified
5. **Generates Markdown**: Outputs a markdown checklist with checkboxes for each requirement

## Example Output

The generated checklist includes:
- Policy ID as the title
- Control sections organized by control ID
- Checklist items (checkboxes) for each assessment requirement
- Modified requirements with rationale and recommendations

See `osps-checklist-example.md` for a complete example output.
