# Simple App Ritual Example

A simple example ritual demonstrating the basic structure and features of Toutago Ritual Grove.

## Purpose

This example ritual shows:
- Basic `ritual.yaml` structure
- Simple questionnaire with conditional questions
- Template rendering with variable substitution
- Protected files management
- Post-install hooks

## Structure

```
simple-app/
├── ritual.yaml          # Ritual manifest
├── templates/           # Template files (not included in this minimal example)
│   ├── main.go.fith
│   └── config.yaml.fith
└── static/             # Static files
    └── README.md
```

## Features Demonstrated

### 1. Questionnaire

The ritual asks for:
- **app_name**: Application name (validated with regex and length constraints)
- **description**: Optional description with default value
- **port**: HTTP port with numeric validation
- **database_type**: Choice among none/postgres/mysql/sqlite
- **database_host**: Conditional field (only if database selected)
- **database_port**: Conditional field (only for postgres)

### 2. Conditional Questions

Shows how to use `condition` to show/hide questions based on other answers:

```yaml
- name: database_host
  condition:
    field: database_type
    not_equals: none

- name: database_port
  condition:
    field: database_type
    equals: postgres
```

### 3. Protected Files

Demonstrates marking files as protected from updates:

```yaml
files:
  protected:
    - config/config.yaml
    - README.md
```

### 4. Post-Install Hooks

Shows running commands after project generation:

```yaml
hooks:
  post_install:
    - go mod tidy
    - go build -o {{app_name}}
```

## Using This Example

### 1. Test the Ritual

```bash
cd examples/simple-app
touta ritual validate
```

### 2. Generate a Project

```bash
touta ritual init simple-app --output /tmp/my-test-app
```

### 3. Answer the Questions

The questionnaire will prompt you for:
1. Application name
2. Description
3. Port number
4. Database selection
5. Database connection details (if applicable)

### 4. Inspect Generated Project

```bash
cd /tmp/my-test-app
ls -la
cat config/config.yaml
```

## Customization

To create your own ritual based on this example:

1. **Copy the structure**:
   ```bash
   cp -r examples/simple-app my-ritual
   cd my-ritual
   ```

2. **Modify ritual.yaml**:
   - Change name, description, author
   - Add/remove questions
   - Update file mappings
   - Customize hooks

3. **Add templates**:
   - Create templates/ directory
   - Add .fith or .tmpl files
   - Use {{variable}} syntax for substitution

4. **Test your ritual**:
   ```bash
   touta ritual validate
   touta ritual init my-ritual --output /tmp/test
   ```

## Key Concepts

### Variable Substitution

Variables from questionnaire answers are available in templates and hooks:
- `{{app_name}}` - User's application name
- `{{port}}` - Selected port number
- `{{database_type}}` - Selected database

### Validation Rules

Questions support various validation:
- `pattern`: Regular expression
- `min_len`, `max_len`: String length constraints
- `min`, `max`: Numeric range
- `required`: Must be provided

### File Types

- **Templates**: Rendered with variable substitution
- **Static**: Copied as-is without modification
- **Protected**: Never overwritten during updates

## Related Examples

- `minimal-app/` - Even simpler example with fewer features
- `../rituals/blog/` - Complex ritual with full application
- `../rituals/basic-site/` - Production-ready simple website

## Learn More

- [Ritual Format Documentation](../../docs/ritual-format.md)
- [Questionnaire Guide](../../docs/questionnaire.md)
- [Template Engine](../../docs/templates.md)
