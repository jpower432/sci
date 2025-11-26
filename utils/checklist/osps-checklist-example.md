
# Implementation Guidance: osps-example-policy

## OSPS-AC-01

**Control:** OSPS-B / OSPS-AC-01

- [ ] **OSPS-AC-01.01**: When a user attempts to access a sensitive resource in the project's version control system, the system MUST require the user to complete a multi-factor authentication process using hardware tokens or passkeys. - Clarified MFA requirements for cloud environments. When a user attempts to access a sensitive resource in the project's version control system, the system MUST require the user to complete a multi-factor authentication process using hardware tokens or passkeys. Recommendation: Implement hardware security keys or passkeys for all administrative access

---

## OSPS-AC-03

**Control:** OSPS-B / OSPS-AC-03

- [ ] **OSPS-AC-03.01**: When a direct commit is attempted on the project's primary branch,
an enforcement mechanism MUST prevent the change from being applied.
 - When a direct commit is attempted on the project's primary branch,
an enforcement mechanism MUST prevent the change from being applied.
 Recommendation: If the VCS is centralized, set branch protection on the primary branch
in the project's VCS. Alternatively, use a decentralized approach,
like the Linux kernel's, where changes are first proposed in another
repository, and merging changes into the primary repository requires a
specific separate act.

- [ ] **OSPS-AC-03.02**: When an attempt is made to delete the project's primary branch,
the version control system MUST treat this as a sensitive activity
and require explicit confirmation of intent.
 - When an attempt is made to delete the project's primary branch,
the version control system MUST treat this as a sensitive activity
and require explicit confirmation of intent.
 Recommendation: Set branch protection on the primary branch in the project's version
control system to prevent deletion.


