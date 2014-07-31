zhash
=====

Bored of type switches when dealing with huge nested maps? zhash is for you!

Create one
```golang
import zhash

func main() {
    hash := zhash.NewHash()
}
```

Then fill it from toml:

```golang
    hash.ReadHash(reader)
```

or initialize from existing `map[string]interface{}`:

```golang
    hash := zhash.HashFromMap(yourFancyMap)
```

And use it through different getters and setters:
```golang
    s, _ := hash.GetString("path", "to", "nested", "item")

    hash.Set("Some new var", "path", "to", "existing", "or", "new", "element")
```


