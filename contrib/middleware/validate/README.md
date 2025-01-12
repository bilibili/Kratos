# Validator Middleware for Kratos Project

This module provides a middleware for Kratos to validate request parameters, using schema defined in `.proto` files.

There used to be a middleware named `Validator` in Kratos, which calls the generated validation functions
from [PGV](https://github.com/bufbuild/protoc-gen-validate) at runtime. Since PGV has been
in [maintenance](https://github.com/bufbuild/protoc-gen-validate/commit/4a8ffc4942463929c4289407cd4b8c8328ff5422), and
recommend using [protovalidate](https://github.com/bufbuild/protovalidate) as an alternative.

That's why we provide a new middleware that uses the schema definitions and validation functions provided by
protovalidate.

protovalidate no longer requires code generation at build time, but for compatibility with existing Kratos
projects, we enable the legacy mode of protovalidate. For most users, no changes are needed to existing code. **But for
users who have manually implemented the Validator interface, you need to migrate the relevant implementation yourself**.