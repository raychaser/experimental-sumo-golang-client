# WARNING

This is an initial **EXPERIMENTAL** attempt at creating both
Sumo Logic API bindings for Golang as well as an actual CLI
to interact with the Sumo Logic service.

Please consider this "my first Go project"––the author is NOT
a professional Go programmer. However, an attempt was made to
follow best practices for building API and CLI clients.

At this point this code is not meant for production use. This
code is **NOT** an official product of Sumo Logic.


# What Is This?

An initial attempt at providing Sumo Logic API bindings for Go.
The code is based on Google's [go-github](https://github.com/google/go-github)
Github API bindings  which the author understood at the time of 
initial creation in July 2020 to be using best practices for a 
API bindings.

Some of the files contain models generated using the OpenAPI 
generator directly from the Sumo Logic OpenAPI definitions, 
copied over artisinally by hand... Ideally, an automated process 
of some sort would continuously update the bindings based on 
re-generating code from the OpenAPI definitions but this is still
work that needs to be done.

Relying more fully on the OpenAPI-based generation of the API
will likely also mean to rely on the generators underlying API
call code to make the HTTPS calls, vs. what is now found in
`sumologic.go`.
