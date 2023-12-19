# Changelog
All notable changes to this project will be documented in this file.

This project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## 0.4.0 (20 Dec 2023)
### Added
- [#63][]: `AnyOf` matcher for values that satisfy at least one matcher.
- [#114][]: `Regex` matcher.

### Fixed
- [#93][]: Ignore interfaces that contain type constraints.
- [#101][]: Fix race condition on Controller.Satisfied.
- [#121][]: Fix paths for windows.
- [#127][]: Propogate -mock_names to Recorder and ReturnCall types.
- [#132][]: Sanitize "any" package names.

[#63]: https://github.com/uber-go/mock/pull/63
[#93]: https://github.com/uber-go/mock/pull/93
[#101]: https://github.com/uber-go/mock/pull/101
[#114]: https://github.com/uber-go/mock/pull/114
[#121]: https://github.com/uber-go/mock/pull/121
[#127]: https://github.com/uber-go/mock/pull/127
[#132]: https://github.com/uber-go/mock/pull/132

Thanks to @favonia @hoonmin @pshopper @davidharrigan @dlsniper @merrett010
@craig65535 @chemidy @tulzke @UnAfraid @JacobOaks @sywhang for their
contributions this release.

## 0.3.0 (15 Sep 2023)
### Added
- [#13][]: Support for embedded generic interfaces.
- [#33][]: `-write_source_comment` for writing the original file or interface names
  in the generated code.
- [#46][]: `-write-generate-directive` for generating go:generate directives into
  the generated mock.
- [#60][]: `Cond` matcher for specifying a conditional matcher as the result of a
  given function.
- [#72][]: `exclude_interfaces` flag for specifying list of interfaces to exclude
  from mock generation.

### Fixed
- [#41][]: A bug with specifying local import name with `-imports` flag.
- [#52][]: A panic that occurs in `gob.Register` when used in conjunction with
  golang/mock.
- [#78][]: `InOrder` can be used with type-safe mocks generated with `-typed` flag.

[#13]: https://github.com/uber-go/mock/pull/13
[#33]: https://github.com/uber-go/mock/pull/33
[#41]: https://github.com/uber-go/mock/pull/41
[#46]: https://github.com/uber-go/mock/pull/46
[#52]: https://github.com/uber-go/mock/pull/52
[#60]: https://github.com/uber-go/mock/pull/60
[#72]: https://github.com/uber-go/mock/pull/72
[#78]: https://github.com/uber-go/mock/pull/78

Thanks to @alexandear, @bcho, @deathiop, @sivchari, @k3forx, @n0trace,
@utgwkk, @ErfanMomeniii, @bcho, @damianopetrungaro, @Tulzke,
and @EstebanOlmedo for their contributions to this release.

## 0.2.0 (06 Jul 2023)
### Added
- `Controller.Satisfied` that lets you check whether all expected calls
  bound to a Controller have been satisfied.
- `NewController` now takes optional `ControllerOption` parameter.
- `WithOverridableExpectations` is a `ControllerOption` that configures
  Controller to override existing expectations upon a new EXPECT().
- `-typed` flag for generating type-safe methods in the generated mock.

## 0.1.0 (29 Jun 2023)

This is a minor version that mirrors the original golang/mock
project that this project originates from.

Any users on golang/mock project should be able to migrate to
this project as-is, and expect exact same set of features (apart
from supported Go versions. See [README](README.md#supported-go-versions)
for more details.
