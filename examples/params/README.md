# Parameter Passing Examples

This example demonstrates the different ways to pass parameters to Guix components.

## Overview

Guix supports multiple parameter passing styles to give you flexibility in how you design your components:

1. **Normal Parameters** - Simple, direct parameter passing
2. **Props Struct** - Structured parameter passing with a dedicated type
3. **Multiple Parameters** - Multiple individual parameters
4. **Auto Props (@props directive)** - Automatic props struct and option functions generation
5. **Variadic Parameters** - Variable number of arguments (like Go's `...` syntax)
6. **Channels** - Reactive data streams (works with @props)

## Parameter Passing Styles

### 1. Normal Parameters

The simplest approach - pass parameters directly as function arguments.

**Definition:**
```go
func SimpleCard(title string, description string) (Component) {
    Div(Class("card")) {
        H2(Class("card-title")) {
            `{title}`
        }
        P(Class("card-description")) {
            `{description}`
        }
    }
}
```

**Generated Constructor:**
```go
func NewSimpleCard(title string, description string) *SimpleCard
```

**Usage:**
```go
SimpleCard("Welcome", "This is a simple card")
```

**When to use:**
- Simple components with 1-3 parameters
- When parameter names are self-explanatory
- When you don't need optional parameters

### 2. Props Struct

Pass all parameters as a single struct. You define the props struct yourself.

**Definition:**
```go
type UserProfileProps struct {
    Username string
    Email    string
    Age      int
}

func UserProfile(props UserProfileProps) (Component) {
    Div(Class("user-profile")) {
        H3 {
            `{props.Username}`
        }
        P {
            `Email: {props.Email}`
        }
        P {
            `Age: {props.Age}`
        }
    }
}
```

**Generated Constructor:**
```go
func NewUserProfile(props UserProfileProps) *UserProfile
```

**Usage:**
```go
UserProfile(UserProfileProps{
    Username: "john_doe",
    Email:    "john@example.com",
    Age:      30,
})
```

**When to use:**
- Components with many parameters
- When you want to reuse the props type elsewhere
- When you want explicit type documentation
- When you need to pass the same props to multiple components

### 3. Multiple Parameters

Similar to normal parameters but with more arguments.

**Definition:**
```go
func ProductCard(productName string, price float64, inStock bool) (Component) {
    Div(Class("product-card")) {
        H3 {
            `{productName}`
        }
        P {
            `Price: ${price}`
        }
    }
}
```

**Generated Constructor:**
```go
func NewProductCard(productName string, price float64, inStock bool) *ProductCard
```

**Usage:**
```go
ProductCard("Laptop", 999.99, true)
```

**When to use:**
- When you have 3-5 clearly distinct parameters
- When parameter order is logical and memorable

### 4. Auto Props (@props directive)

Use the `@props` directive to automatically generate a props struct, option functions, and a variadic option constructor.

**Definition:**
```go
@props func AutoCard(title string, subtitle string, highlighted bool) (Component) {
    Div(Class("auto-card")) {
        H2 {
            `{title}`
        }
        H4 {
            `{subtitle}`
        }
    }
}
```

**Generated Code:**
```go
// Props struct
type AutoCardProps struct {
    Title       string
    Subtitle    string
    Highlighted bool
}

// Option type
type AutoCardOption func(*AutoCard)

// Option functions
func WithTitle(v string) AutoCardOption {
    return func(c *AutoCard) {
        c.Title = v
    }
}

func WithSubtitle(v string) AutoCardOption {
    return func(c *AutoCard) {
        c.Subtitle = v
    }
}

func WithHighlighted(v bool) AutoCardOption {
    return func(c *AutoCard) {
        c.Highlighted = v
    }
}

// Constructor
func NewAutoCard(opts ...AutoCardOption) *AutoCard
```

**Usage:**
```go
AutoCard(
    WithTitle("Auto-generated"),
    WithSubtitle("Using @props directive"),
    WithHighlighted(true)
)
```

**When to use:**
- When you want optional parameters
- When you want to make some parameters optional in usage
- When you want a functional options pattern
- When you're building a component library and want a clean API

### 5. Variadic Parameters

Use variadic parameters when you want to accept a variable number of arguments of the same type.

**Definition:**
```go
func MessageList(messages ...string) (Component) {
    Div(Class("message-list")) {
        H3 {
            "Messages"
        }
        for _, msg in messages {
            Div(Class("message-item")) {
                `{msg}`
            }
        }
    }
}
```

**Generated Constructor:**
```go
func NewMessageList(messages ...string) *MessageList
```

**Usage:**
```go
MessageList("Hello", "World", "From", "Guix")
```

**When to use:**
- When you need a variable number of items of the same type
- When building list or collection components
- When the number of items isn't known at compile time

### 6. Channels with @props

Combine channels (reactive data streams) with the @props directive for components that update in response to data changes.

**Definition:**
```go
@props func LiveCounter(count chan int) (Component) {
    currentCount := <-count
    Div(Class("live-counter")) {
        Span(Class("count-value")) {
            `Count: {currentCount}`
        }
    }
}
```

**Generated Constructor:**
```go
func NewLiveCounter(opts ...LiveCounterOption) *LiveCounter
```

**Usage:**
```go
counterChan := make(chan int, 10)
LiveCounter(WithCount(counterChan))

// Update the component by sending to the channel
counterChan <- 42
```

**When to use:**
- When your component needs to react to changing data
- When building real-time UIs
- When you want automatic re-rendering on data updates

## Comparison Table

| Style | Constructor Signature | Best For | Flexibility |
|-------|----------------------|----------|-------------|
| Normal Params | `New(arg1, arg2)` | 1-3 simple parameters | Low |
| Props Struct | `New(props Props)` | Many parameters, reusable types | Medium |
| Multiple Params | `New(arg1, arg2, arg3, arg4)` | 3-5 distinct parameters | Low |
| Auto Props | `New(opts ...Option)` | Optional params, clean API | High |
| Variadic | `New(items ...Type)` | Variable number of same-type items | Medium |
| Channels | `New(opts ...Option)` | Reactive, real-time data | High |

## Running the Example

```bash
# Build the example
guix build examples/params

# The generated Go code will be in:
# - examples/params/params_gen.go
# - examples/params/app_gen.go
```

## Key Takeaways

1. **Without `@props`**: The constructor accepts the exact same parameters as defined in the component function
2. **With `@props`**: The constructor uses a variadic options pattern with auto-generated option functions
3. **Variadic parameters**: Use `...` syntax just like in Go for variable-length argument lists
4. **Props structs**: Define manually when you want full control over the type
5. **Choose based on your use case**: Simple components → normal params, complex components → @props or props struct

## Best Practices

- Use **normal parameters** for simple, 1-3 parameter components
- Use **@props** when you want flexibility and optional parameters
- Use **props struct** when you need to reuse the props type or have many parameters
- Use **variadic** when you need variable-length lists of the same type
- Use **channels** with **@props** for reactive components

The choice depends on your specific needs, team preferences, and the complexity of your components!
