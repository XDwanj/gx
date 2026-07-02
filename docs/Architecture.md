# 架构设计：Tree-sitter 能力调查与 `gx` 的 `kind` 设计

## 目的

本文档记录以下内容：

- `gx` 当前支持语言的真实 Tree-sitter 能力边界
- grammar 能力与当前 `gx` 抽取覆盖之间的差异
- `gx` 符号 `kind` 的设计约束
- `gx` 长期推荐采用的 `kind` 模型

本文档用于指导未来对以下部分的修改：

- `internal/language/language.go`
- `internal/language/queries.go`
- `internal/index/types.go`
- `gx symbols --kind` 与 `gx definition --kind` 的用户文档

## 范围

### 当前活跃语言

截至当前代码库，`gx` 在 `internal/language/language.go` 中注册了以下语言：

- `bash`
- `c`
- `cpp`
- `go`
- `java`
- `lua`
- `python`
- `ruby`
- `rust`
- `swift`
- `typescript`
- `zig`

这也是当前 `kind` 设计真正面向的目标语言集合。

- `bash`
- `c`
- `cpp`
- `go`
- `java`
- `lua`
- `python`
- `ruby`
- `rust`
- `swift`
- `typescript`
- `zig`

## 重要边界：runtime grammar 不等于 `gx` 覆盖范围

`gx` 使用 gotreesitter 的内置 grammar registry 加载语法，但 `gx` 的公开语言范围仍由
`internal/lang` 和 `internal/language` 中注册的语言决定。gotreesitter registry
包含更多 grammar，不代表这些语言都已经成为 `gx` 的公开支持面。

这意味着，`kind` 设计必须由 `gx` 当前注册语言的综合 Tree-sitter 能力面来驱动，而不能只看 gotreesitter registry 中还提供了哪些额外 grammar。

## 方法

本次调查基于三层信息源：

1. 当前语言注册与抽取行为：
   - `internal/language/language.go`
   - `internal/language/queries.go`
2. gotreesitter 内置 grammar registry 与对应 AST 节点形状
3. 本地 fixture 与 `gx` 当前公开输出行为

本次调查明确区分三个概念：

- **Grammar capability**：grammar 能否用稳定的语法节点，或稳定的父子结构，表达该语法概念
- **Reasonable queryability**：`gx` 能否在不依赖脆弱启发式规则的前提下稳定抽取它
- **Current `gx` coverage**：当前 `queries.go` 是否真的已经把它抽出来

这个区分很重要，因为 `gx` 目前不少缺口并不是 grammar 做不到，而只是 query 覆盖还没补齐。

## 能力标签说明

下文表格使用以下标签：

- `native`：grammar 有直接对应的语法节点，或者存在非常稳定的语法形式
- `derived`：grammar 能支持这个概念，但通常需要依赖上下文或语言特定模式，而不是一个单独的直接节点
- `weak`：技术上可表示，但噪音较大，或语义过于不稳定，不适合作为 `gx` 默认的符号来源
- `none`：这个语言里没有有意义的原生语法概念与之对应

## 调查总结

### 语言清单

| 语言         | Grammar 来源            | 说明                                                                           |
| ------------ | ----------------------- | ------------------------------------------------------------------------------ |
| `bash`       | gotreesitter registry   | 声明面很窄，主要是函数与赋值                                                   |
| `c`          | gotreesitter registry   | 对函数、结构体、枚举、typedef 有较强结构支持                                   |
| `cpp`        | gotreesitter registry   | 对函数、方法、类、结构体、枚举、命名空间有较强支持                             |
| `go`         | gotreesitter registry   | 对函数、方法、常量、变量、结构体、接口、命名类型有较强支持                     |
| `java`       | gotreesitter registry   | 对类、接口、枚举、方法、模块、枚举成员有较强支持                               |
| `lua`        | gotreesitter registry   | 支持函数、表方法语法、赋值；没有原生类模型                                     |
| `python`     | gotreesitter registry   | 支持类、函数、赋值、装饰定义；方法需要依赖上下文判断                           |
| `ruby`       | gotreesitter registry   | 支持 class/module/method/constant 语法；没有原生 enum/interface 语法           |
| `rust`       | gotreesitter registry   | 对结构体、枚举、trait、常量、模块、variant、类型别名有较强支持                 |
| `swift`      | gotreesitter registry   | 对 class/struct/enum/protocol/extension/typealias/method/property 语法支持很强 |
| `typescript` | gotreesitter registry   | 对 class、interface、enum、namespace/module、变量、方法、字段支持很强          |
| `zig`        | gotreesitter registry   | 支持函数、`const`/`var` 声明、container 类型与 error set                       |

### 能力矩阵

这张表回答的问题是：如果 `gx` 想支持一套更干净的跨语言 `kind` 模型，那么哪些保留语言能真正用 Tree-sitter 为这些类别提供支撑？

| 语言         | `fn`   | `method` | `const` | 包级/顶层 `var` | `struct` | `enum`  | `class` | `interface` | `module` | 枚举值  | `type`  |
| ------------ | ------ | -------- | ------- | --------------- | -------- | ------- | ------- | ----------- | -------- | ------- | ------- |
| `bash`       | native | none     | weak    | weak            | none     | none    | none    | none        | none     | none    | none    |
| `c`          | native | none     | derived | derived         | native   | native  | none    | none        | none     | native  | native  |
| `cpp`        | native | native   | derived | derived         | native   | native  | native  | derived     | native   | native  | native  |
| `go`         | native | native   | native  | native          | native   | derived | none    | native      | weak     | derived | native  |
| `java`       | none   | native   | native  | weak            | none     | native  | native  | native      | native   | native  | weak    |
| `lua`        | native | native   | weak    | weak            | none     | none    | none    | none        | none     | none    | none    |
| `python`     | native | derived  | derived | derived         | none     | derived | native  | derived     | none     | weak    | derived |
| `ruby`       | none   | native   | derived | weak            | none     | none    | native  | none        | native   | none    | weak    |
| `rust`       | native | derived  | native  | none            | native   | native  | none    | native      | native   | native  | native  |
| `swift`      | native | native   | native  | derived         | native   | native  | native  | native      | derived  | derived | native  |
| `typescript` | native | native   | native  | native          | none     | native  | native  | native      | native   | native  | native  |
| `zig`        | native | none     | native  | native          | native   | native  | none    | none        | weak     | derived | derived |

### 当前 `gx` 抽取快照

这张表总结的是：当前 `queries.go` 在保留语言集合上到底已经抽取了哪些内容。

| 语言         | 当前已抽取 kind                                                                   | 高价值缺口                                                 |
| ------------ | --------------------------------------------------------------------------------- | ---------------------------------------------------------- |
| `bash`       | `fn`                                                                              | 赋值名是有意不纳入的                                       |
| `c`          | `fn`, `struct`, `enum`, `type`                                                    | 类常量名字、枚举值                                         |
| `cpp`        | `fn`, `method`, `struct`, `class`, `enum`, `module`, `type`                       | 缺枚举值                                                   |
| `go`         | `fn`, `method`, `const`, `struct`, `interface`, `type`                            | 枚举式值通过 `const` 支持，但仍没有独立 `enum`             |
| `java`       | `class`, `method`, `interface`, `enum`, `module`, `const`                         | 仍没有单独的 `type` 细分                                   |
| `lua`        | `fn`, `method`                                                                    | 赋值名是有意不纳入的                                       |
| `python`     | `const`, `class`, `fn`                                                            | 类内方法没有单独分成 `method`；装饰形式未完整暴露          |
| `ruby`       | `method`, `class`, `module`                                                       | 缺常量                                                     |
| `rust`       | `fn`, `method`, `struct`, `enum`, `interface`, `module`, `type`, `const`          | `trait` 已归并到 `interface`，值级名字统一并入 `const`     |
| `swift`      | `fn`, `method`, `const`, `struct`, `enum`, `class`, `interface`, `module`, `type` | 当前覆盖最强，主要只剩更细的常量/属性区分问题              |
| `typescript` | `fn`, `method`, `class`, `interface`, `type`, `enum`, `module`, `const`           | 字段级值符号仍未纳入默认抽取                               |
| `zig`        | `fn`, `struct`, `enum`                                                            | 缺稳定的 `const` 抽取，缺值级名字，`type` 也没有被有效使用 |

## 各语言调查记录

### `bash`

- Tree-sitter 支持 `function_definition`
- Tree-sitter 也支持 `variable_assignment`
- 这个语言没有稳定的 module、class、interface、enum 或 struct 模型

架构含义：

- `bash` 应继续作为 `gx` 的最小语言支持集存在
- `fn` 是唯一明显高价值的默认 kind
- 基于赋值的名字噪音太大，不适合广泛纳入默认符号抽取

### `c`

- 存在直接支持的 `function_definition`
- 存在直接支持的 `struct_specifier`、`enum_specifier` 与 `type_definition`
- 枚举值在结构上也是可表示的
- 变量与类似常量的声明在语法上存在，但语言本身并没有一个特别干净、跨项目稳定的 `const` 符号类别

架构含义：

- `fn`、`struct`、`enum`、`type` 都是很合适的类别
- `const` 后续可以支持，但它在这里的价值明显不如 Go 或 Rust

### `cpp`

- 对自由函数与成员函数存在直接支持
- 对 `struct`、`class`、`enum` 与 `type_definition` 存在直接支持
- 命名空间也是稳定的语法特性
- 枚举成员可以作为值级声明被表示

架构含义：

- `cpp` 强烈支持保留 `struct`、`class`、`enum`、`method`、`module`
- 当前把 `struct_specifier` 映射成 `class` 的做法过于粗糙

### `go`

- 存在直接支持的：
  - `function_declaration`
  - `method_declaration`
  - `const_declaration` / `const_spec`
  - `var_declaration` / `var_spec`
  - `type_spec`
  - `struct_type`
  - `interface_type`
- Go 没有原生 `enum`，但通过命名类型 + `const` + `iota` 建模 enum-like 值是惯用写法，而且语法上可查询

架构含义：

- 当前 `gx` 对 Go 的支持明显偏弱
- `go` 应支持：
  - `fn`
  - `method`
  - `const`
  - `struct`
  - `interface`
  - `type`
- 包级 `var` 合理地可以并入 `const`，用于导航
- enum-like 值也应该可搜索，并映射为 `const`

### `java`

- 存在直接支持的：
  - `class_declaration`
  - `interface_declaration`
  - `enum_declaration`
  - `method_declaration`
  - `module_declaration`
  - `constant_declaration`
  - `enum_constant`
- Java 没有一个独立于 method 的有意义 `fn` 类别

架构含义：

- `java` 强烈支持 `class`、`interface`、`enum`、`method`、`module`、`const`
- Java 的枚举值应当可搜索，并且最适合归入 `const`

### `lua`

- 存在直接支持的：
  - `function_declaration`
  - 点号索引函数名
  - 方法索引函数名
  - 赋值语句
- 这个语言没有原生的 class/interface/enum/module 模型

架构含义：

- `fn` 和 `method` 是仅有的高价值稳定 kind
- 基于赋值的抽取虽可行，但太弱，不适合作为默认符号索引

### `python`

- 存在直接支持的：
  - `function_definition`
  - `class_definition`
  - `assignment`
  - `decorated_definition`
- 方法不是独立节点类型，而是 class 内部的函数
- enum 与 interface-like 行为更多来自库约定，而不是原生语法

架构含义：

- `class` 与 `fn` 是明显合适的类别
- `method` 可以通过祖先节点关系来支持
- 模块级赋值可以支持 `const`
- Python 不值得继续增加更多公开 kind

### `ruby`

- 存在直接支持的：
  - `method`
  - `singleton_method`
  - `class`
  - `module`
  - `constant`
- Ruby 没有原生 enum 或 interface 语法

架构含义：

- `method`、`class`、`module`、`const` 是最有价值的类别
- Ruby 常量完全可以被索引，无需再发明新的 kind

### `rust`

- 存在直接支持的：
  - `function_item`
  - `const_item`
  - `static_item`
  - `struct_item`
  - `enum_item`
  - `enum_variant`
  - `trait_item`
  - `mod_item`
  - `type_item`
- Rust 的方法是通过 `impl` 等 item body 上下文推导出来的

架构含义：

- Rust 强烈支持：
  - `fn`
  - `const`
  - `struct`
  - `enum`
  - 通过把 `trait` 映射为 `interface` 来支持抽象契约
  - `module`
  - `type`
- Rust 枚举 variant 应当可搜索，并映射为 `const`
- 一旦已经有 `interface`，公开的 `trait` kind 就没有必要再单独存在

### `swift`

- 存在直接支持的：
  - class、struct、enum、actor、extension、protocol、typealias
  - 顶层函数
  - 方法
  - protocol method declaration
  - 属性
  - constant declaration
- Swift 没有单独的 trait 模型；protocol 就是天然的抽象契约

架构含义：

- Swift 强烈支持：
  - `fn`
  - `method`
  - `const`
  - `struct`
  - `enum`
  - `class`
  - `interface`
  - `type`
- 如果 `gx` 决定公开 `extension`，它可以映射为 `module`

### `typescript`

- 存在直接支持的：
  - `function_declaration`
  - `method_definition`
  - `class_declaration`
  - `interface_declaration`
  - `type_alias_declaration`
  - `enum_declaration`
  - `module`
  - `internal_module`
  - `lexical_declaration`
  - `variable_declaration`
  - `public_field_definition`
  - enum body 与值级成员
- TypeScript 是最适合做广覆盖 `gx kind` 设计的语言之一

架构含义：

- TypeScript 强烈支持：
  - `fn`
  - `method`
  - `const`
  - `class`
  - `interface`
  - `enum`
  - `module`
  - `type`
- enum 成员应当可被索引，并映射为 `const`

### `zig`

- 存在直接支持的：
  - `FnProto`
  - `VarDecl`
  - `struct`、`enum`、`union` 的 container declaration
  - `ErrorSetDecl`
  - `const` 与 `var` 形式
- Zig 没有原生 class 或 interface 模型

架构含义：

- Zig 支持：
  - `fn`
  - `const`
  - `struct`
  - `enum`
  - `type`
- 如果 `gx` 不想引入新 kind，`union` 可以继续归到 `struct`
- `ErrorSetDecl` 可以继续归到 `enum`

## 当前 `gx` 覆盖问题

重构前，`gx` 的 `kind` 集合是：

- `fn`
- `method`
- `struct`
- `enum`
- `trait`
- `type`
- `const`
- `class`
- `interface`
- `module`
- `event`

核心问题不只是“数量多”，而是“设计与覆盖不匹配”。

### 1. 有些 kind 太稀疏

- `event` 基本只由 `solidity` 支撑
- `trait` 只由 Rust 支撑，而 Rust trait 完全可以自然映射到 `interface`

一旦移除 `solidity` 和 `elixir`，这两个 kind 都很难继续作为公开、稳定的一等分类存在。

### 2. 一些高价值 kind 在最重要的语言里反而支持不足

最典型的例子就是 Go。

Go grammar 明明已经支持：

- `const`
- 包级 `var`
- `struct`
- `interface`

但当前 `gx` 对 Go 只暴露了：

- `fn`
- `method`
- `type`

这会让 Go 的使用体验比保留语言集合实际允许的能力弱很多。

### 3. `type` 当前被过度滥用

在多种语言里，`gx` 之所以把很多东西放进 `type`，并不是因为语言本身缺乏更细的结构，而只是因为当前抽取太粗。

Go 是最明显的例子：

- `type User struct {}` 应该是 `struct`，而不只是 `type`
- `type Store interface {}` 应该是 `interface`，而不只是 `type`

### 4. 值级符号基本缺失

保留语言里有不少高价值的值级符号，用户经常希望直接按名字搜索：

- Go `const` 块里的 enum-like 值
- Java `enum_constant`
- Rust `enum_variant`
- TypeScript enum member
- Ruby constant

当前 `gx` 基本都没有覆盖这些。

## `gx kind` 的设计约束

`gx` 是一个导航工具，不是语言本体论建模工具。

因此，正确的 `kind` 设计目标既不是：

- “保留每种语言的全部语义区别”

也不是：

- “把所有东西都压扁到一个通用 type 桶里”

真正应该追求的是：

- 保留那些在多个保留语言里都稳定存在的区别
- 避免保留只由单一已计划删除语言或边缘语言支撑的 kind
- 搜索体验优先于语义教科书式精确
- 对外公开的 kind 数量要稳定、可学习

## 推荐的 `kind` 模型

### 最终推荐的公开 kind 集合

推荐对外公开的 `kind` 集合是：

- `fn`
- `method`
- `const`
- `struct`
- `enum`
- `class`
- `interface`
- `module`
- `type`

### 推荐移除的 kind

- 移除 `event`
- 移除 `trait`

### 为什么这组最合适

#### 保留 `fn`

- 几乎所有保留语言都需要
- 大多数保留语言都有直接语法支撑

#### 保留 `method`

- 在 Go、Java、Lua、Ruby、Rust、Swift、TypeScript、C++ 中都很有价值

#### 保留 `const`

- 用来统一吸收：
  - 字面常量声明
  - 枚举成员
  - Go 包级 `var`
  - Ruby 常量风格名字

这是一种最实用的办法，可以让用户搜索命名值，而不需要继续新增高噪音的 `var` 或 `enum_member` kind。

#### 保留 `struct`

- C、C++、Go、Rust、Swift、Zig 都强烈支撑它存在

#### 保留 `enum`

- C、C++、Java、Rust、Swift、TypeScript、Zig 都明显需要它

#### 保留 `class`

- C++、Java、Ruby、Swift、TypeScript 都会用到

#### 保留 `interface`

- Go、Java、Swift、TypeScript 都需要它
- Rust `trait` 直接并入这里也很自然

#### 保留 `module`

- 对 Rust module、Ruby module、C++ namespace、Java module、TypeScript module/namespace 仍然有价值

#### 保留 `type`

- 它仍然需要作为兜底类别，用来承载：
  - 类型别名
  - opaque 或抽象命名类型
  - 无法更自然落入其他类别的命名类型

但 `type` 不应再承担“垃圾桶”角色。那些本来就明显属于 `struct`、`interface` 或 `enum` 的东西，不应该再被压进去。

## 推荐的映射规则

### 跨语言统一规则

- 自由函数映射到 `fn`
- 绑定到类型、receiver、对象、协议的方法映射到 `method`
- 命名值映射到 `const`
- 枚举成员映射到 `const`
- Go 包级 `var` 映射到 `const`
- 结构体风格的命名容器类型映射到 `struct`
- 枚举声明映射到 `enum`
- 类风格的命名对象类型映射到 `class`
- interface、protocol、Rust trait 统一映射到 `interface`
- module 与 namespace 映射到 `module`
- 类型别名以及其余未归类命名类型映射到 `type`

### 明确不建议公开的 kind

- `trait`
  - 一旦已经有 `interface`，它就过于 Rust 专属
- `event`
  - 在移除 `solidity` 后支撑面太窄
- `var`
  - 作为公开跨语言类别噪音太大
- `enum_member`
  - 概念上有意义，但更适合直接暴露为 `const`

## 推荐的语言映射调整

### `go`

推荐支持：

- `function_declaration` -> `fn`
- `method_declaration` -> `method`
- `type_spec` 且其 `type` 为 `struct_type` -> `struct`
- `type_spec` 且其 `type` 为 `interface_type` -> `interface`
- 其他 `type_spec` -> `type`
- `const_spec` -> `const`
- 包级 `var_spec` -> `const`

这是当前代码库里价值最高、也最明显缺失的一块。

### `rust`

推荐支持：

- `const_item` 与 `static_item` -> `const`
- `enum_variant` -> `const`
- `trait_item` -> `interface`

### `typescript`

推荐支持：

- 保留当前的 `fn`、`method`、`class`、`interface`、`enum`、`module`、`type`
- 额外补充值级抽取：
  - `const` 声明
  - enum member
  - 如果证明有足够价值，可选地加入稳定 class field

### `java`

推荐支持：

- 保留当前的 `class`、`method`、`interface`、`enum`
- 补充：
  - `module_declaration` -> `module`
  - `constant_declaration` -> `const`
  - `enum_constant` -> `const`

### `cpp`

推荐支持：

- 保留 `fn`、`method`、`enum`、`type`
- 把 `struct_specifier` 从 `class` 改成 `struct`
- 保留 `class_specifier` 为 `class`
- 考虑加入 `namespace_definition` -> `module`
- 考虑把 enum 成员映射为 `const`

### `ruby`

推荐支持：

- 保留 `method`、`class`、`module`
- 补充稳定的常量抽取 -> `const`

### `python`

推荐支持：

- 保留 `class` 与顶层 `const`
- class 内函数分类为 `method`
- 顶层函数保留为 `fn`
- 通过解包 `decorated_definition` 支持装饰形式

### `swift`

推荐支持：

- 保留当前广覆盖能力
- 只有在更复杂的逻辑能明显改善搜索质量时，才进一步区分“真正的常量声明”和“一般属性声明”
- protocol 继续统一映射到 `interface`

### `zig`

推荐支持：

- 保留 `fn`
- 容器支撑的命名类型继续映射到：
  - `struct`
  - `enum`
- 稳定的命名值映射到 `const`
- `ErrorSetDecl` 继续归到 `enum`

## 最终结论

当前 `gx` 的问题并不是抽象意义上的 “kind 太多”。
更深层的问题是：

- 有些低价值 kind 虽然公开存在，但支撑太弱
- 一些高价值 kind 在保留语言里反而支持不完整
- `type` 正在替代 query 缺口，而不是作为刻意设计的兜底类别存在

正确的方向应该是：

1. 适度缩减公开 kind，移除 `event` 与 `trait`
2. 强化保留语言中的高价值支持，尤其是 Go
3. 将值级命名统一纳入 `const`
4. 让 `type` 回到真正的兜底角色，而不是继续承担“大杂烩”职责

面向保留语言集合，推荐的稳定公开 `kind` 集合是：

- `fn`
- `method`
- `const`
- `struct`
- `enum`
- `class`
- `interface`
- `module`
- `type`

这是当前 `gx` 最合适的模型，因为它：

- 与保留语言集合的能力边界匹配度高
- 保留了足够有用的搜索表达力
- 避免了 kind 爆炸
- 也避免把有价值的差异全部压平
