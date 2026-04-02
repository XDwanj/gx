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
  body: (_)) @definition.class

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
  name: (type_identifier) @name
  body: (_)) @definition.class

(class_specifier
  name: (type_identifier) @name) @definition.class

(enum_specifier
  name: (type_identifier) @name) @definition.enum

(type_definition
  declarator: (type_identifier) @name) @definition.type
`

const elixirQuery = `
(call
  target: (identifier) @_keyword
  (arguments (alias) @name)
  (#any-of? @_keyword "defmodule" "defprotocol" "defimpl")) @definition.module

(call
  target: (identifier) @_keyword
  (arguments
    [(identifier) @name
     (call target: (identifier) @name)
     (binary_operator left: (call target: (identifier) @name))])
  (#any-of? @_keyword "def" "defp" "defmacro" "defmacrop" "defguard" "defguardp" "defdelegate")) @definition.function

(unary_operator
  operand: (call
    target: (identifier) @_keyword
    (arguments
      (binary_operator
        left: (identifier) @name)))
  (#any-of? @_keyword "type" "typep" "opaque")) @definition.type

(unary_operator
  operand: (call
    target: (identifier) @_keyword
    (arguments
      (binary_operator
        left: (call target: (identifier) @name)))
    (#eq? @_keyword "callback"))) @definition.method
`

const goQuery = `
(function_declaration
  name: (identifier) @name) @definition.function

(method_declaration
  name: (field_identifier) @name) @definition.method

(type_spec
  name: (type_identifier) @name) @definition.type
`

const javaQuery = `
(class_declaration
  name: (identifier) @name) @definition.class

(method_declaration
  name: (identifier) @name) @definition.method

(interface_declaration
  name: (identifier) @name) @definition.interface

(enum_declaration
  name: (identifier) @name) @definition.enum
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
(module (assignment left: (identifier) @name) @definition.constant)

(class_definition
  name: (identifier) @name) @definition.class

(function_definition
  name: (identifier) @name) @definition.function
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
    name: (type_identifier) @name) @definition.class

(enum_item
    name: (type_identifier) @name) @definition.class

(union_item
    name: (type_identifier) @name) @definition.class

(type_item
    name: (type_identifier) @name) @definition.class

(declaration_list
    (function_item
        name: (identifier) @name) @definition.method)

(function_item
    name: (identifier) @name) @definition.function

(trait_item
    name: (type_identifier) @name) @definition.interface

(mod_item
    name: (identifier) @name) @definition.module

(macro_definition
    name: (identifier) @name) @definition.macro
`

const solidityQuery = `
(contract_declaration
  name: (identifier) @name) @definition.class

(interface_declaration
  name: (identifier) @name) @definition.interface

(library_declaration
  name: (identifier) @name) @definition.module

(function_definition
  name: (identifier) @name) @definition.function

(struct_declaration
  name: (identifier) @name) @definition.class

(enum_declaration
  name: (identifier) @name) @definition.enum

(event_definition
  name: (identifier) @name) @definition.event
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

(module
  name: (identifier) @name) @definition.module

(lexical_declaration
  (variable_declarator
    name: (identifier) @name
    value: (arrow_function))) @definition.function

(variable_declaration
  (variable_declarator
    name: (identifier) @name
    value: (arrow_function))) @definition.function
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
            "struct")))))) @definition.class

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
            "union")))))) @definition.class

(Decl
  (VarDecl
    (IDENTIFIER) @name
    (ErrorUnionExpr
      (SuffixExpr
        (ErrorSetDecl))))) @definition.enum
`
