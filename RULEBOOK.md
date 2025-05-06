# Cursor IDE Rulebook

## Code Style and Formatting
1. Always use consistent indentation (4 spaces for Go)
2. Follow Go's standard formatting conventions
3. Use meaningful variable and function names
4. Add comments for complex logic or important functions
5. Keep functions focused and single-purpose

## Error Handling
1. Always handle errors explicitly
2. Use descriptive error messages
3. Log errors with appropriate context
4. Return meaningful HTTP status codes
5. Provide clear error responses to clients

## Testing
1. Write tests for all new functionality
2. Include both positive and negative test cases
3. Test edge cases and error conditions
4. Keep tests focused and readable
5. Use table-driven tests where appropriate

## HTTP Handler Guidelines
1. Always set appropriate Content-Type headers
2. Handle all HTTP methods properly
3. Validate input data
4. Return appropriate status codes
5. Follow RESTful principles

## Storage Operations
1. Use context for all storage operations
2. Handle concurrent access properly
3. Validate data before storage
4. Clean up resources properly
5. Use appropriate error handling

## Logging
1. Use structured logging
2. Include relevant context in log messages
3. Log at appropriate levels (debug, info, error)
4. Don't log sensitive information
5. Use consistent log format

## Security
1. Validate all input data
2. Don't expose sensitive information
3. Use appropriate HTTP headers
4. Handle authentication properly
5. Follow security best practices

## Performance
1. Use efficient data structures
2. Minimize memory allocations
3. Handle large payloads properly
4. Use appropriate buffering
5. Consider concurrent operations

## Documentation
1. Document public APIs
2. Include usage examples
3. Document configuration options
4. Keep documentation up to date
5. Use clear and concise language

## Code Organization
1. Keep related code together
2. Use appropriate package structure
3. Minimize dependencies
4. Follow Go's package conventions
5. Keep files focused and manageable

## Version Control
1. Write clear commit messages
2. Keep commits focused and atomic
3. Use appropriate branching strategy
4. Review code before committing
5. Keep history clean and meaningful

## Debugging
1. Use appropriate logging levels
2. Include relevant context in logs
3. Use debug tools effectively
4. Document known issues
5. Handle edge cases properly

## Maintenance
1. Keep dependencies up to date
2. Remove unused code
3. Fix technical debt
4. Monitor performance
5. Regular code reviews

## Best Practices
1. Follow Go's idioms
2. Use standard library when possible
3. Write maintainable code
4. Consider future extensibility
5. Keep code simple and clear 
