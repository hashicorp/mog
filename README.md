# mog

[![Build Status](https://github.com/hashicorp/mog/workflows/ci/badge.svg)](https://github.com/hashicorp/mog/actions)

mog is a [Go](http://www.golang.org) code generation tool for converting API types into Core types

The current use cases for this tool is to automatically generate conversion
routines from Consul API types to Consul Core Types.

## Installation

If you wish to build mog you'll need Go version 1.18+ installed.

Please check your installation with:

```
go version
```

To install Mog:

```
  go install github.com/hashicorp/mog@latest
```

## Documentation

Mog is configured with annotations in comments on the source side in two places: structs and fields.

### Struct Annotations

Structs are opted into `mog` handling by adding an annotation. After the
standard struct comment you can add a single line by itself like:

    // mog annotation:

And then you can follow it with whitespace-delimited `key=value` directives:

| Key             | Type     | Meaning                                                                                 |
| --------------- | -------- | --------------------------------------------------------------------------------------- |
| `target`        | required | Fully qualified identifier for the other side of this `mog` conversion mapping.         |
| `output`        | required | Name of generated output file to put the generated functions into.                      |
| `name`          | required | Suffix for generated bidirectional conversion functions. Those two functions will be `<StructName><To|From><NameSuffix>`. |
| `ignore-fields` | optional | Comma-delimited list of source fields that should be ignored for conversion mapping.    |
| `func-from`     | optional | TBD |
| `func-to`       | optional | TBD |

#### Example

    // Blah blah blah regular comment.
    //
    // mog annotation:
    //
    // target=github.com/hashicorp/consul/agent/structs.Something
    // output=service.gen.go
    // name=Structs
    // ignore-fields=Kind,Name,RaftIndex,EnterpriseMeta
    message Something {

### Struct Field Annotations

After the standard struct field comment you can OPTIONALLY add a single line by
itself like this to adjust the automatic conversion routines:

    // mog:<SPACE>

The rest of that line should be a series of whitespace-delimited `key=value`
directives:

| Key         | Meaning                                                                                                                                  |
| ------------| ---------------------------------------------------------------------------------------------------------------------------------------- |
| `target`    | Field name for the other side of this `mog` conversion mapping. If unspecified a field with the same name is assumed.                    |
| `pointer`   | _reserved and unused_                                                                                                                    |
| `func-from` | Name of function to use to do the copying/conversion from TARGET to SOURCE. The signature should take one argument and return one value. |
| `func-to`   | Name of function to use to do the copying/conversion to TARGET from SOURCE. The signature should take one argument and return one value. |

#### Examples

    // things that require manual work
    mog: func-to=MapHeadersToStructs func-from=NewMapHeadersFromStructs
    mog: func-to=CheckTypesToStructs func-from=NewCheckTypesFromStructs
    mog: func-to=EnterpriseMetaTo func-from=EnterpriseMetaFrom
    mog: func-to=RaftIndexToStructs func-from=NewRaftIndexFromStructs
    mog: func-to=intentionActionToStructs func-from=intentionActionFromStructs

    // type alias helpers
    mog: func-to=uint func-from=uint32
    mog: func-to=int func-from=int32
    mog: func-to=NodeIDType func-from=string
    mog: func-to=CheckIDType func-from=string
    mog: func-to=structs.ServiceKind func-from=string
    mog: func-to=structs.MeshGatewayMode func-from=string
    mog: func-to=structs.ProxyMode func-from=string

    // protobuf types
    mog: func-to=structs.DurationFromProto func-from=structs.DurationToProto
    mog: func-to=structs.TimeFromProto func-from=structs.TimeToProto
    mog: func-to=TimePtrFromProto func-from=TimePtrToProto

    // unfortunate protobuf camel-casing help (protoc will uppercase the first x)
    mog: target=EnforcingConsecutive5xx
