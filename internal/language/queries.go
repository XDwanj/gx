package language

const bashQuery = `
(function_definition
  name: (word) @name) @definition.function
`

const cQuery = `
(function_definition
  declarator: (function_declarator
    declarator: (identifier) @name)) @definition.function

(function_definition
  declarator: (pointer_declarator
    declarator: (function_declarator
      declarator: (identifier) @name))) @definition.function

(struct_specifier
  name: (type_identifier) @name
  body: (_)) @definition.struct

(enum_specifier
  name: (type_identifier) @name) @definition.enum

(type_definition
  declarator: (type_identifier) @name) @definition.type
`

const cppQuery = `
(function_definition
  declarator: (function_declarator
    declarator: (identifier) @name)) @definition.function

(function_definition
  declarator: (pointer_declarator
    declarator: (function_declarator
      declarator: (identifier) @name))) @definition.function

(function_definition
  declarator: (function_declarator
    declarator: (field_identifier) @name)) @definition.method

(function_definition
  declarator: (function_declarator
    declarator: (qualified_identifier
      name: (identifier) @name))) @definition.method

(struct_specifier
  name: [(type_identifier) (qualified_identifier)] @name
  body: (_)) @definition.struct

(class_specifier
  name: [(type_identifier) (qualified_identifier)] @name) @definition.class

(enum_specifier
  name: (type_identifier) @name) @definition.enum

(namespace_definition
  name: [(namespace_identifier) (nested_namespace_specifier)] @name) @definition.module

(type_definition
  declarator: (type_identifier) @name) @definition.type
`

const goQuery = `
(function_declaration
  name: (identifier) @name) @definition.function

(method_declaration
  name: (field_identifier) @name) @definition.method

(source_file
  (type_declaration
    (type_spec
      name: (type_identifier) @name
      type: (struct_type)) @definition.struct))

(source_file
  (type_declaration
    (type_spec
      name: (type_identifier) @name
      type: (interface_type)) @definition.interface))

(source_file
  (type_declaration
    (type_spec
      name: (type_identifier) @name) @definition.type))

(source_file
  (type_declaration
    (type_alias
      name: (type_identifier) @name) @definition.type))

(source_file
  (const_declaration
    (const_spec
      name: (identifier) @name) @definition.constant))

(source_file
  (const_declaration
    (const_spec
      name: (identifier) @name
      value: (_)) @definition.constant))

(source_file
  (var_declaration
    (var_spec
      name: (identifier) @name) @definition.constant))

(source_file
  (var_declaration
    (var_spec_list
      (var_spec
        name: (identifier) @name) @definition.constant)))

(source_file
  (var_declaration
    (var_spec
      name: (identifier) @name
      value: (_)) @definition.constant))
`

const javaQuery = `
(module_declaration
  name: [(identifier) (scoped_identifier)] @name) @definition.module

(class_declaration
  name: (identifier) @name) @definition.class

(method_declaration
  name: (identifier) @name) @definition.method

(interface_declaration
  name: (identifier) @name) @definition.interface

(enum_declaration
  name: (identifier) @name) @definition.enum

(constant_declaration
  declarator: (variable_declarator
    name: (identifier) @name)) @definition.constant

(field_declaration
  (modifiers
    "final")
  declarator: (variable_declarator
    name: (identifier) @name)) @definition.constant

(enum_constant
  name: (identifier) @name) @definition.constant
`

const kotlinQuery = `
(class_declaration
  (modifiers
    (class_modifier) @class_modifier)
  "class"
  name: (identifier) @name) @definition.enum
  (#eq? @class_modifier "enum")

(class_declaration
  "class"
  name: (identifier) @name) @definition.class

(class_declaration
  "interface"
  name: (identifier) @name) @definition.interface

(object_declaration
  name: (identifier) @name) @definition.class

(function_declaration
  name: (identifier) @name) @definition.function

(property_declaration
  (variable_declaration
    (identifier) @name)) @definition.constant

(property_declaration
  (multi_variable_declaration
    (variable_declaration
      (identifier) @name))) @definition.constant

(enum_entry
  (identifier) @name) @definition.constant

(type_alias
  type: (identifier) @name) @definition.type
`

const luaQuery = `
(function_declaration
  name: (identifier) @name) @definition.function

(function_declaration
  name: (dot_index_expression
    field: (identifier) @name)) @definition.function

(function_declaration
  name: (method_index_expression
    method: (identifier) @name)) @definition.method
`

const pythonQuery = `
(assignment
  left: (identifier) @name) @definition.constant

(assignment
  left: (pattern_list
    (identifier) @name)) @definition.constant

(assignment
  left: (tuple_pattern
    (identifier) @name)) @definition.constant

(assignment
  left: (list_pattern
    (identifier) @name)) @definition.constant

(class_definition
  name: (identifier) @name) @definition.class

(function_definition
  name: (identifier) @name) @definition.function
`

const protobufQuery = `
(message
  (message_name) @name) @definition.struct

(enum
  (enum_name) @name) @definition.enum

(service
  (service_name) @name) @definition.interface

(rpc
  (rpc_name) @name) @definition.method
`

const rubyQuery = `
(method
  name: (_) @name) @definition.method

(singleton_method
  name: (_) @name) @definition.method

(class
  name: (constant) @name) @definition.class

(module
  name: (constant) @name) @definition.module
`

const rustQuery = `
(struct_item
  name: (type_identifier) @name) @definition.struct

(enum_item
  name: (type_identifier) @name) @definition.enum

(union_item
  name: (type_identifier) @name) @definition.struct

(type_item
  name: (type_identifier) @name) @definition.type

(declaration_list
  (function_item
    name: (identifier) @name) @definition.method)

(function_item
  name: (identifier) @name) @definition.function

(trait_item
  name: (type_identifier) @name) @definition.interface

(const_item
  name: (identifier) @name) @definition.constant

(static_item
  name: (identifier) @name) @definition.constant

(enum_variant
  name: (identifier) @name) @definition.constant

(mod_item
  name: (identifier) @name) @definition.module

(macro_definition
  name: (identifier) @name) @definition.macro
`

const swiftQuery = `
; --- Type declarations ---

(class_declaration
  "class"
  name: (type_identifier) @name) @definition.class

(class_declaration
  "struct"
  name: (type_identifier) @name) @definition.struct

(class_declaration
  "enum"
  name: (type_identifier) @name) @definition.enum

(class_declaration
  "actor"
  name: (type_identifier) @name) @definition.class

(class_declaration
  "extension"
  name: _ @name) @definition.module

(protocol_declaration
  name: (type_identifier) @name) @definition.interface

(typealias_declaration
  name: (type_identifier) @name) @definition.type

; --- Methods (functions inside type bodies) ---

(class_body
  (function_declaration
    name: (simple_identifier) @name) @definition.method)

(enum_class_body
  (function_declaration
    name: (simple_identifier) @name) @definition.method)

(protocol_body
  (protocol_function_declaration
    name: (simple_identifier) @name) @definition.method)

; --- Init / Deinit ---

(class_body
  (init_declaration
    name: _ @name) @definition.method)

(enum_class_body
  (init_declaration
    name: _ @name) @definition.method)

(class_body
  (deinit_declaration
    "deinit" @name) @definition.method)

; --- Subscripts ---

(class_body
  (subscript_declaration
    "subscript" @name) @definition.method)

(enum_class_body
  (subscript_declaration
    "subscript" @name) @definition.method)

; --- Properties ---

(class_body
  (property_declaration
    name: (pattern
      bound_identifier: (simple_identifier) @name)) @definition.constant)

(enum_class_body
  (property_declaration
    name: (pattern
      bound_identifier: (simple_identifier) @name)) @definition.constant)

(protocol_body
  (protocol_property_declaration
    name: (pattern
      bound_identifier: (simple_identifier) @name)) @definition.constant)

; --- Top-level functions ---

(function_declaration
  name: (simple_identifier) @name) @definition.function
`

const typeScriptQuery = `
(function_declaration
  name: (identifier) @name) @definition.function

(class_declaration
  name: (type_identifier) @name) @definition.class

(method_definition
  name: (property_identifier) @name) @definition.method

(interface_declaration
  name: (type_identifier) @name) @definition.interface

(type_alias_declaration
  name: (type_identifier) @name) @definition.type

(enum_declaration
  name: (identifier) @name) @definition.enum

(internal_module
  name: (identifier) @name) @definition.module

(lexical_declaration
  (variable_declarator
    name: (identifier) @name
    value: (arrow_function))) @definition.function

(variable_declaration
  (variable_declarator
    name: (identifier) @name
    value: (arrow_function))) @definition.function

(lexical_declaration
  (variable_declarator
    name: (identifier) @name)) @definition.constant

(variable_declaration
  (variable_declarator
    name: (identifier) @name)) @definition.constant

(enum_assignment
  name: [(property_identifier) (string) (number)] @name) @definition.constant

(enum_body
  name: [(property_identifier) (string) (number)] @name) @definition.constant
`

const zigQuery = `
(Decl
  (FnProto
    (IDENTIFIER) @name)) @definition.function

(Decl
  (VarDecl
    (IDENTIFIER) @name
    (ErrorUnionExpr
      (SuffixExpr
        (ContainerDecl
          (ContainerDeclType
            "struct")))))) @definition.struct

(Decl
  (VarDecl
    (IDENTIFIER) @name
    (ErrorUnionExpr
      (SuffixExpr
        (ContainerDecl
          (ContainerDeclType
            "enum")))))) @definition.enum

(Decl
  (VarDecl
    (IDENTIFIER) @name
    (ErrorUnionExpr
      (SuffixExpr
        (ContainerDecl
          (ContainerDeclType
            "union")))))) @definition.struct

(Decl
  (VarDecl
    (IDENTIFIER) @name
    (ErrorUnionExpr
      (SuffixExpr
        (ErrorSetDecl))))) @definition.enum
`
