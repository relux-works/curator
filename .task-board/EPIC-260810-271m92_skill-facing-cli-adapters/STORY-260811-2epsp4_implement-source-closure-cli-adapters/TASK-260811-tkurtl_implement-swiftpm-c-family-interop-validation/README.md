# TASK-260811-tkurtl: implement-swiftpm-c-family-interop-validation

## Description
Implement the source and trust validation for SwiftPM mixed Swift and C-family boundaries. Spec trace: .spec/skill-facing-cli-source-closure.md Current delivery scope SwiftPM-supported C family; Source closure invariant items 4-6; Vendored compiled artifact prohibition. Accepted input: TASK-260810-zddzh7 Source, header, module closure and System libraries sections.

## Scope
Model separate Swift and Clang capture targets; classify C, C++, Objective-C, and Objective-C++ sources; parse and confine custom and generated module maps; verify compiler-observed headers and modules; represent C ABI, C++ interoperability, Objective-C runtime, linker, SDK, sysroot, and FFI declarations in capture while requiring exact platform, Swift or Clang toolchain, SDK, targets, uses-tool, and selected system or interop bindings in SelectionBinding. Validate selected-product reachability and reject duplicate, dangling, wrong-kind, incompatible, or capture-replacing bindings, mixed Swift and C-family sources in one target, unsafe flags, active plugins or macros, binary targets, and untrusted system libraries.

## Acceptance Criteria
Supported Swift-to-C and restricted C++, Objective-C, and Objective-C++ fixtures expose explicit capture declarations plus exact platform, toolchain, SDK, FFI, and build-order bindings. Every header, module, framework, library, and SDK read resolves to admitted source or one C0-selected binding node; no target or tool edge is hidden in node payloads or capture selection state. CGP05 target-binding and CGN03, CGN09, CGN15 cases plus H01-H08, S02-S09, and P01-P09 produce the accepted pass or stable fail-closed diagnostics before extension execution.
