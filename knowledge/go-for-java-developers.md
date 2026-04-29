# Go for Java Backend Developers

A crash course highlighting key differences with practical exercises.

---

## 1. Variables & Types

**The Java Way:**
```java
String name = "Alice";
final int MAX_SIZE = 100; // constant
var count = 10; // type inference (Java 10+)
```

**The Go Way:**
```go
var name string = "Alice"  // explicit
name := "Alice"            // shorthand (inferred, only inside functions)
const MaxSize = 100        // constant (PascalCase = exported)

// Multiple declarations
var a, b int = 1, 2
x, y := 3, "hello"
```

**Key Differences:**
- `:=` is your new best friend for local variables
- Types come *after* the name (`name string`, not `String name`)
- `const` is for compile-time constants
- Zero values: `0`, `""`, `false`, `nil` (no null pointer exceptions for unassigned vars)

---

### Exercises

**Exercise 1.1:** Convert this Java to Go:
```java
final double TAX_RATE = 0.08;
int quantity = 5;
double price = 19.99;
double total = quantity * price * (1 + TAX_RATE);
```

**Exercise 1.2:** What are the zero values for these Go variables?
```go
var count int
var name string
var active bool
var user *User
```
Predict them, then write a program to print them.

**Exercise 1.3:** Fix this code:
```go
package main

func main() {
    var message = "Hello"    // line 1
    message := "World"       // line 2
    const MAX = 100          // line 3
    MAX = 200                // line 4
}
```

---

## 2. Functions

**The Java Way:**
```java
public int add(int a, int b) {
    return a + b;
}

// Multiple return values? Use a wrapper class or pass mutable objects.
```

**The Go Way:**
```go
// Simple function
func add(a, b int) int {
    return a + b
}

// Multiple return values (very common!)
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

// Named return values (use sparingly)
func rectangleProps(width, height float64) (area, perimeter float64) {
    area = width * height
    perimeter = 2 * (width + height)
    return // naked return - returns the named variables
}
```

**Key Differences:**
- No `public`/`private` - capitalized = exported (public), lowercase = package-private
- Multiple return values replace throwing exceptions or returning result objects
- No function overloading (use different names or variadic functions)

---

### Exercises

**Exercise 2.1:** Write a function `calculate` that takes two `float64` numbers and returns their sum, difference, product, and quotient (in that order).

**Exercise 2.2:** Write a function `findUser` that returns a `User` and an `error`. Return an error if `id` is <= 0.

```go
type User struct {
    ID   int
    Name string
}
```

**Exercise 2.3:** Java has `System.out.println(String.format("Hello %s", name))`. Go uses:
```go
func fmt.Sprintf(format string, a ...interface{}) string
```
Write a variadic function `concat` that takes any number of strings and returns them concatenated.

---

## 3. Structs vs Classes

**The Java Way:**
```java
public class User {
    private String name;
    private int age;
    
    public User(String name, int age) {
        this.name = name;
        this.age = age;
    }
    
    public String getName() { return name; }
    public void setName(String name) { this.name = name; }
}
```

**The Go Way:**
```go
type User struct {
    Name string  // exported (public)
    age  int     // unexported (private to package)
}

// Constructor pattern (not required, but common)
func NewUser(name string, age int) *User {
    return &User{Name: name, age: age}
}

// Method (value receiver - gets a copy)
func (u User) IsAdult() bool {
    return u.age >= 18
}

// Method (pointer receiver - can modify)
func (u *User) HaveBirthday() {
    u.age++
}

// Usage
user := NewUser("Alice", 25)
fmt.Println(user.IsAdult())  // true
user.HaveBirthday()
```

**Key Differences:**
- No classes, only structs with methods "attached"
- No inheritance (prefer composition via embedding)
- No constructors - use factory functions like `NewUser`
- Decide carefully: value receiver (immutable) vs pointer receiver (mutable)

---

### Exercises

**Exercise 3.1:** Create a `Product` struct with exported fields `Name`, `Price`, and an unexported field `sku`. Add a constructor `NewProduct`.

**Exercise 3.2:** Add methods:
- `Discount(percent float64)` - pointer receiver, reduces price
- `GetSKU()` - value receiver, returns the SKU

**Exercise 3.3:** Go uses **composition** instead of inheritance. Convert this Java inheritance:

```java
class Animal {
    String name;
    void speak() { }
}

class Dog extends Animal {
    void speak() { System.out.println("Woof"); }
}
```

Into Go using struct embedding.

---

## 4. Interfaces (Implicit Implementation!)

**The Java Way:**
```java
interface Speaker {
    void Speak();
}

class Dog implements Speaker {
    public void Speak() { System.out.println("Woof"); }
}
```

**The Go Way:**
```go
type Speaker interface {
    Speak() string
}

type Dog struct{}

func (d Dog) Speak() string {
    return "Woof"
}

// Dog implicitly implements Speaker - no "implements" keyword!

func MakeSound(s Speaker) {
    fmt.Println(s.Speak())
}

// Usage
dog := Dog{}
MakeSound(dog)  // Works! No explicit declaration needed.
```

**Key Differences:**
- **Implicit implementation** - if you have the methods, you implement the interface
- Small interfaces are preferred (`io.Reader`, `io.Writer` with just 1 method!)
- Enables loose coupling and easy mocking for tests

---

### Exercises

**Exercise 4.1:** Define a `PaymentProcessor` interface with a `Pay(amount float64) error` method. Implement it for `CreditCard` and `PayPal` structs.

**Exercise 4.2:** Write a function `ProcessPayment(p PaymentProcessor, amount float64)` that uses the interface.

**Exercise 4.3:** The empty interface `interface{}` (or `any` in Go 1.18+) accepts any type. Write a function `PrintType(v interface{})` that prints the dynamic type of `v` using `fmt.Printf("%T\n", v)`.

---

## 5. Concurrency: Goroutines vs Threads

**The Java Way:**
```java
// Thread per task - expensive!
Thread t = new Thread(() -> {
    System.out.println("Running in thread");
});
t.start();

// Or use ExecutorService
ExecutorService executor = Executors.newFixedThreadPool(10);
executor.submit(() -> { ... });
```

**The Go Way:**
```go
// Goroutine - lightweight (2KB stack, grows/shrinks)
go func() {
    fmt.Println("Running in goroutine")
}()

// Channels for communication (think BlockingQueue)
messages := make(chan string)  // unbuffered

// Send
go func() {
    messages <- "hello"  // blocks until receiver ready
}()

// Receive
msg := <-messages  // blocks until message available
fmt.Println(msg)

// Buffered channel (async up to capacity)
buffered := make(chan int, 10)
```

**Key Differences:**
- Goroutines are **cheap** - you can have millions
- **Don't share memory to communicate; communicate to share memory**
- Channels are typed and safe
- `select` for multiplexing channels (like `Selector` in Java NIO)

```go
select {
case msg1 := <-ch1:
    fmt.Println("From ch1:", msg1)
case msg2 := <-ch2:
    fmt.Println("From ch2:", msg2)
case <-time.After(5 * time.Second):
    fmt.Println("Timeout!")
}
```

---

### Exercises

**Exercise 5.1:** Write a function that launches 3 goroutines, each printing its number (0, 1, 2). Notice the output is unpredictable. Fix it using `sync.WaitGroup`.

**Exercise 5.2:** Create a `worker` function that receives jobs from a channel and sends results to another channel. Simulate processing by sleeping 100ms.

**Exercise 5.3:** Write a `fanIn` function that multiplexes multiple channels into one. Use `select` to read from multiple input channels.

---

## 6. Error Handling (No Exceptions!)

**The Java Way:**
```java
try {
    User user = userService.findById(id);
    process(user);
} catch (UserNotFoundException e) {
    logger.error("User not found", e);
} catch (Exception e) {
    logger.error("Unexpected", e);
} finally {
    cleanup();
}
```

**The Go Way:**
```go
user, err := userService.FindByID(id)
if err != nil {
    // Handle error - usually return it or log and return
    return fmt.Errorf("finding user: %w", err)  // wrap error
}
process(user)

// No finally - use defer!
file, err := os.Open("file.txt")
if err != nil {
    return err
}
defer file.Close()  // executes when function returns
```

**Key Differences:**
- Errors are **values**, not exceptions
- Explicit error checking with `if err != nil`
- `defer` for cleanup (runs LIFO when function returns)
- `panic`/`recover` exist but are for truly exceptional cases (like `OutOfMemoryError`)

---

### Exercises

**Exercise 6.1:** Write a function `safeDivide(a, b int) (int, error)` that returns an error when dividing by zero. Test it.

**Exercise 6.2:** Demonstrate `defer` LIFO order:
```go
defer fmt.Println("1")
defer fmt.Println("2")
defer fmt.Println("3")
```
What will print?

**Exercise 6.3:** Create a custom error type `ValidationError` with a `Field` property. Return it from a `validateUser` function and use `errors.As` to check for it.

---

## 7. Collections

**The Java Way:**
```java
List<String> list = new ArrayList<>();
list.add("a");
list.add("b");
String first = list.get(0);

Map<String, Integer> map = new HashMap<>();
map.put("key", 100);
```

**The Go Way:**
```go
// Slice (like ArrayList, but more flexible)
list := []string{"a", "b"}  // literal
list = append(list, "c")    // returns new slice (may reallocate)

first := list[0]      // access
last := list[len(list)-1]

// Slice operations (no built-in filter/map - write loops or use generics in 1.18+)
subset := list[1:3]   // slice[low:high], half-open range

// Map
m := make(map[string]int)
m["key"] = 100

value, exists := m["key"]  // comma ok idiom
if exists {
    fmt.Println(value)
}

// Delete
delete(m, "key")
```

**Key Differences:**
- Slices are references to underlying arrays
- `append` may create a new backing array
- Maps return optional second value for existence check
- No generics before 1.18; now available but loops are still common

---

### Exercises

**Exercise 7.1:** Create a slice of 5 integers. Use a loop to double each value in place.

**Exercise 7.2:** Write a function `countWords(text string) map[string]int` that returns word frequencies.

**Exercise 7.3:** Implement a function `filter(nums []int, predicate func(int) bool) []int` that returns a new slice containing only elements matching the predicate.

---

## 8. Quick Reference Table

| Java | Go |
|------|-----|
| `class` | `struct` + methods |
| `interface` (explicit) | `interface` (implicit) |
| `extends` | embedding (composition) |
| `public`/`private` | Capitalized/`lowercase` |
| `null` | `nil` |
| `try-catch` | `if err != nil` |
| `Thread` | `goroutine` |
| `ArrayList` | `slice` |
| `HashMap` | `map` |
| `final` | `const` |
| `List<User>` | `[]User` |
| `Map<K,V>` | `map[K]V` |
| `ExecutorService` | `go` keyword + channels |

---

## 9. Why `go run .` is Slow on First Run

`go run` **compiles** your code before executing it. The first run is slow because Go is compiling your program to machine code. Subsequent runs are fast because Go caches the build artifacts.

```
First run:   Compile -> Cache -> Run  (slow)
Second run:  Check cache -> Run     (fast - skip compilation)
```

### What `go run` Actually Does

```bash
go run .     # Equivalent to: go build -o /tmp/exe . && /tmp/exe
```

Unlike Java which runs bytecode through the JVM (interpreter/JIT), Go compiles directly to **native machine code** (like C/C++).

### Go vs Java Flow

| | **Go** | **Java** |
|---|---|---|
| **Source** | `.go` files | `.java` files |
| **Compile** | Native binary (one-time, slow) | Bytecode `.class` files |
| **Run** | Direct execution (fast) | JVM interprets/JIT compiles |
| **First run** | Slow (compilation) | Fast (just bytecode) |
| **Subsequent** | Fast (cached binary) | Similar (JVM still works) |

### Where is the Cache?

Go stores build cache in:

```bash
# Location
$GOPATH/pkg/mod        # Module cache
$GOCACHE              # Build cache (usually ~/go/pkg or %LOCALAPPDATA%\go-build on Windows)

# Check yours
go env GOCACHE
```

### How to Verify This

Create a test file:

```go
// main.go
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
```

Run with timing:

```bash
# First run - includes compilation
time go run .
# real    0m0.5xxs  (slow)

# Second run - cached
time go run .
# real    0m0.0xxs  (fast!)
```

Use `-x` flag to see what Go is doing:

```bash
go run -x .    # Shows all compile/link steps
```

### Want It Even Faster?

If you want to skip the compilation check entirely, **build once and run the binary:**

```bash
# Build (slow)
go build -o myapp .

# Run directly (fastest - no Go toolchain involved)
./myapp
```

This is similar to Java's `javac` then `java`, but Go produces a standalone executable.

### Summary

| Command | What Happens | Speed |
|---------|-------------|-------|
| `go run .` (1st) | Compile + run | Slow |
| `go run .` (2nd+) | Use cache + run | Fast |
| `go build` + `./exe` | Compile once, run many | Fastest |

This is normal behavior - Go trades initial compile time for **runtime performance**. Your final binary runs at native speed without any VM overhead!
