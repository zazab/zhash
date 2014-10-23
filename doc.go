/*
	Package zhash gives you the tool to operate huge map[string]interface{} with pleasure.

	Creating Hash

	There are two methods to create new Hash. First, you can create Hash from
	existing map[string]interface{} using HashFromMap(m) function:
		m := map[string]interface{}{
			"field": "value",
			"Id": 10,
		}

		h := zhash.HashFromMap(m)

	Or you can create an empty hash by NewHash() function. Then you can fill it
	via Set, or you can set any function fiting the Unmarshaller type, and read
	your hash from any reader (file or bytes.Buffer, for example).
		h := zhash.NewHash()
		h.SetUnmarshallerFunc(json.Unmarshal)

		h.ReadHash(fd)

	Accessing data

	So, you have your hash. How can you access it's data? It's simple --- use
	Get<Type> for getting single items, Get<Type>Slice for getting slices,
	Set for changing items, Delete for deleting childs of nested (or not)
	maps, and Append<Type>Slice for appending slices.

	Setting data

	Set make no difference on what was there before setting new value. So,
	you can easily replace any map with int, and loose all underlying data.
	Also Set creates all needed parents if needed, and replaces any found
	element in the way by map[string]interface{}. So be double careful with Set.

	Appending slices

	Append<Type>Slice will succeed if Get<Type>Slice return no err, or err is
	not found error. Append<Type>Slice replaces original slice, so, for example,
	if original slice "some.slice" was []interface{} containing only ints,
	and you do AppendIntSlice, after append "some.slice" would become []int64.

*/
package zhash
