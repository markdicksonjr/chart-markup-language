# Contributing to Chart Markup Language

This document provides guidelines and information for contributors to the Chart Markup Language project.

## How to Contribute

### Reporting Issues

Before creating an issue, please:
1. Check if the issue has already been reported
2. Use the issue template provided
3. Include as much detail as possible (grammar examples, expected behavior, etc.)

### Suggesting Enhancements

We welcome suggestions for:
- New drawing types
- Additional styling options
- New indicator support
- Grammar improvements
- Documentation enhancements

### Code Contributions

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Make your changes** following the style guidelines below
4. **Test your changes** thoroughly
5. **Commit your changes**: `git commit -m 'Add amazing feature'`
6. **Push to the branch**: `git push origin feature/amazing-feature`
7. **Open a Pull Request**

## Development Guidelines

### Grammar Changes

When modifying the EBNF grammar:
- Update the grammar file (`chart-markup-language.ebnf`)
- Update the README.md with examples
- Add test cases in the examples directory
- Update the CHANGELOG.md

### Documentation

- Keep the README.md up to date
- Add examples for new features
- Use clear, concise language
- Include practical examples

### Examples

When adding new examples:
- Place them in the `examples/` directory
- Use descriptive filenames
- Include comments explaining the example
- Test that examples are valid according to the grammar

## Style Guidelines

### EBNF Grammar
- Use consistent indentation (4 spaces)
- Group related rules together
- Add comments for complex rules
- Use descriptive rule names

### Documentation
- Use Markdown formatting consistently
- Include code examples in fenced code blocks
- Use descriptive headings
- Keep line length reasonable (80-100 characters)

## Testing

Before submitting changes:
1. Verify grammar changes with an EBNF parser
2. Test examples manually
3. Ensure documentation is accurate
4. Check for typos and formatting issues

## Community Guidelines

- Be respectful and constructive
- Help others learn and improve
- Focus on the project's goals
- Follow the code of conduct

## Questions?

If you have questions about contributing:
- Open a discussion in the GitHub repository
- Check existing issues and discussions
- Review the documentation

Thank you for contributing to Chart Markup Language.
