# Security

## Security Features

promptc implements multiple layers of security to protect against common attack vectors.

### Path Traversal Protection

**Vulnerability**: Attackers could use `../` sequences in import names to read arbitrary files.

**Mitigation**:
- `validateImportName()` blocks any import containing `..`, `/`, `\`, or drive letters
- All paths are cleaned with `filepath.Clean()` before use
- `loadFromFilesystem()` verifies resolved paths stay within allowed directories
- Double-checks after resolving symlinks

**Tests**: `TestPathTraversalPrevention` - 11 test cases

### Symlink Attack Protection

**Vulnerability**: Symlinks could point outside allowed directories.

**Mitigation**:
- Symlinks are detected with `os.Lstat()`
- `filepath.EvalSymlinks()` resolves the target
- Target path is verified to be within the allowed directory
- Rejected if symlink points outside allowed bounds

**Tests**: `TestSymlinkSecurity`

### File Size Limits

**Vulnerability**: Large files could cause memory exhaustion (YAML bombs, DoS).

**Mitigation**:
- Prompt files limited to 1MB
- Library files limited to 10MB  
- Size checked before reading content

**Tests**: `TestFileSizeLimit`

### Input Validation

**Vulnerability**: Malformed input could cause crashes or unexpected behavior.

**Mitigation**:
- Import names validated with `validateImportName()`
- Only alphanumeric characters, dots, hyphens, and underscores allowed
- Empty strings rejected
- YAML parsing errors handled gracefully

**Tests**: `TestValidImportNameFunction` - 13 test cases

### YAML Bomb Protection

**Vulnerability**: Malicious YAML could expand to consume excessive memory.

**Mitigation**:
- File size limits prevent billion laughs attacks
- Standard YAML parser without custom recursion

## Security Best Practices

### For Users

1. **Review custom libraries** before adding to ~/.prompts/
2. **Validate imports** - be cautious with third-party prompt libraries
3. **Use project-local libraries** for sensitive/proprietary prompts
4. **Keep promptc updated** for latest security patches

### For Contributors

1. **Never bypass validation** - always use `validateImportName()`
2. **Check all file operations** for path traversal risks
3. **Add tests** for any new file operations
4. **Document security implications** of changes

## Reporting Security Issues

If you discover a security vulnerability, please:

1. **DO NOT** open a public issue
2. Email: security@yourproject.com (or use GitHub Security Advisories)
3. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

## Security Audit History

- 2025-01-XX: Initial security hardening
  - Path traversal prevention
  - File size limits
  - Symlink protection
  - Input validation
  - 24 security tests added

## Known Limitations

- Embedded libraries (go:embed) are inherently safe and not subject to the same checks
- Local filesystem access is granted to configured directories (~/.prompts, ./prompts)
- No sandboxing of YAML content itself (trusts yaml.v3 parser)

## Security Checklist for New Features

- [ ] Validate all user input
- [ ] Check for path traversal risks
- [ ] Implement size limits where applicable
- [ ] Add security tests
- [ ] Document security implications
- [ ] Consider symlink attacks
- [ ] Handle errors securely (no information leakage)
