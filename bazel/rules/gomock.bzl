# The MIT License (MIT)
# Copyright © 2018 Jeff Hodges <jeff@somethingsimilar.com>

# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the “Software”), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:

# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.

# THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
# THE SOFTWARE.

# The rules in this files are still under development. Breaking changes are planned.
# DO NOT USE IT.

load("@bazel_skylib//lib:paths.bzl", "paths")
load("@io_bazel_rules_go//go/private:common.bzl", "GO_TOOLCHAIN", "GO_TOOLCHAIN_LABEL")
load("@io_bazel_rules_go//go/private:context.bzl", "go_context")
load("@io_bazel_rules_go//go/private:providers.bzl", "GoArchive", "GoInfo")
load("@io_bazel_rules_go//go/private/rules:wrappers.bzl", go_binary = "go_binary_macro")

_MOCKGEN_TOOL = Label("@org_uber_go_mock//mockgen")
_MOCKGEN_MODEL_LIB = Label("@org_uber_go_mock//mockgen/model")

def _gomock_source_impl(ctx):
    go = go_context(ctx, include_deprecated_properties = False)

    # In Source mode, it's not necessary to pass through a library, as the only thing we use it for is setting up
    # the relative file locations. Forcing users to pass a library makes it difficult in the case where a mock should
    # be included as part of that same library, as it results in a dependency loop (GoMock -> GoInfo -> GoMock).
    # Allowing users to pass an importpath directly bypasses this issue.
    # See the test case in //tests/extras/gomock/source_with_importpath for an example.
    importpath = ctx.attr.source_importpath if ctx.attr.source_importpath else ctx.attr.library[GoInfo].importmap

    # create GOPATH and copy source into GOPATH
    go_path_prefix = "gopath"
    source_relative_path = paths.join("src", importpath, ctx.file.source.basename)
    source = ctx.actions.declare_file(paths.join("gopath", source_relative_path))

    # trim the relative path of source to get GOPATH
    gopath = source.path[:-len(source_relative_path)]
    ctx.actions.run_shell(
        outputs = [source],
        inputs = [ctx.file.source],
        command = "mkdir -p {0} && cp -L {1} {0}".format(source.dirname, ctx.file.source.path),
        mnemonic = "GoMockSourceCopyFile",
    )

    # passed in source needs to be in gopath to not trigger module mode
    args = ["-source", source.path]

    args, needed_files = _handle_shared_args(ctx, args)

    if len(ctx.attr.aux_files) > 0:
        aux_files = []
        for target, pkg in ctx.attr.aux_files.items():
            f = target.files.to_list()[0]
            aux = ctx.actions.declare_file(paths.join(go_path_prefix, "src", pkg, f.basename))
            ctx.actions.run_shell(
                outputs = [aux],
                inputs = [f],
                command = "mkdir -p {0} && cp -L {1} {0}".format(aux.dirname, f.path),
                mnemonic = "GoMockSourceCopyFile",
            )
            aux_files.append("{0}={1}".format(pkg, aux.path))
            needed_files.append(aux)
        args += ["-aux_files", ",".join(aux_files)]

    sdk = go.sdk

    inputs_direct = needed_files + [source]
    inputs_transitive = [sdk.tools, sdk.headers, sdk.srcs]

    # We can use the go binary from the stdlib for most of the environment
    # variables, but our GOPATH is specific to the library target we were given.
    ctx.actions.run_shell(
        outputs = [ctx.outputs.out],
        inputs = depset(inputs_direct, transitive = inputs_transitive),
        tools = [
            ctx.file.mockgen_tool,
            sdk.go,
        ],
        toolchain = GO_TOOLCHAIN_LABEL,
        command = """
            export GOPATH=$(pwd)/{gopath} &&
            export GOROOT=$(pwd)/{goroot} &&
            export PATH=$GOROOT/bin:$PATH &&
            {cmd} {args} > {out}
        """.format(
            gopath = gopath,
            goroot = sdk.root_file.dirname,
            cmd = "$(pwd)/" + ctx.file.mockgen_tool.path,
            args = " ".join(args),
            out = ctx.outputs.out.path,
            mnemonic = "GoMockSourceGen",
        ),
        env = {
            # GOCACHE is required starting in Go 1.12
            "GOCACHE": "./.gocache",
            # gomock runs in the special GOPATH environment
            "GO111MODULE": "off",
        },
    )

_gomock_source = rule(
    _gomock_source_impl,
    attrs = {
        "library": attr.label(
            doc = "The target the Go library where this source file belongs",
            providers = [GoInfo],
            mandatory = False,
        ),
        "source_importpath": attr.string(
            doc = "The importpath for the source file. Alternative to passing library, which can lead to circular dependencies between mock and library targets.",
            mandatory = False,
        ),
        "source": attr.label(
            doc = "A Go source file to find all the interfaces to generate mocks for. See also the docs for library.",
            mandatory = False,
            allow_single_file = True,
        ),
        "out": attr.output(
            doc = "The new Go file to emit the generated mocks into",
            mandatory = True,
        ),
        "aux_files": attr.label_keyed_string_dict(
            default = {},
            doc = "A map from auxilliary Go source files to their packages.",
            allow_files = True,
        ),
        "package": attr.string(
            doc = "The name of the package the generated mocks should be in. If not specified, uses mockgen's default.",
        ),
        "self_package": attr.string(
            doc = "The full package import path for the generated code. The purpose of this flag is to prevent import cycles in the generated code by trying to include its own package. This can happen if the mock's package is set to one of its inputs (usually the main one) and the output is stdio so mockgen cannot detect the final output package. Setting this flag will then tell mockgen which import to exclude.",
        ),
        "imports": attr.string_dict(
            doc = "Dictionary of name-path pairs of explicit imports to use.",
        ),
        "mock_names": attr.string_dict(
            doc = "Dictionary of interface name to mock name pairs to change the output names of the mock objects. Mock names default to 'Mock' prepended to the name of the interface.",
            default = {},
        ),
        "copyright_file": attr.label(
            doc = "Optional file containing copyright to prepend to the generated contents.",
            allow_single_file = True,
            mandatory = False,
        ),
        "mockgen_tool": attr.label(
            doc = "The mockgen tool to run",
            default = _MOCKGEN_TOOL,
            allow_single_file = True,
            executable = True,
            cfg = "exec",
            mandatory = False,
        ),
        "mockgen_args": attr.string_list(
            doc = "Additional arguments to pass to the mockgen tool",
            mandatory = False,
        ),
        "_go_context_data": attr.label(
            default = "@io_bazel_rules_go//:go_context_data",
        ),
    },
    toolchains = [GO_TOOLCHAIN],
)

def gomock(name, out, library = None, source_importpath = "", source = None, interfaces = [], package = "", self_package = "", aux_files = {}, mockgen_tool = _MOCKGEN_TOOL, mockgen_args = [], imports = {}, copyright_file = None, mock_names = {}, **kwargs):
    """Calls [mockgen](https://github.com/uber-go/mock) to generates a Go file containing mocks from the given library.

    If `source` is given, the mocks are generated in source mode; otherwise in archive mode.

    Args:
        name: the target name.
        out: the output Go file name.
        library: the Go library to look into for the interfaces (archive mode) or source (source mode). If running in source mode, you can specify source_importpath instead of this parameter.
        source_importpath: the importpath for the source file. Alternative to passing library, which can lead to circular dependencies between mock and library targets. Only valid for source mode.
        source: a Go file in the given `library`. If this is given, `gomock` will call mockgen in source mode to mock all interfaces in the file.
        interfaces: a list of interfaces in the given `library` to be mocked in archive mode.
        package: the name of the package the generated mocks should be in. If not specified, uses mockgen's default. See [mockgen's -package](https://github.com/uber-go/mock#flags) for more information.
        self_package: the full package import path for the generated code. The purpose of this flag is to prevent import cycles in the generated code by trying to include its own package. See [mockgen's -self_package](https://github.com/uber-go/mock#flags) for more information.
        aux_files: a map from source files to their package path. This only needed when `source` is provided. See [mockgen's -aux_files](https://github.com/uber-go/mock#flags) for more information.
        mockgen_tool: the mockgen tool to run.
        mockgen_args: additional arguments to pass to the mockgen tool.
        imports: dictionary of name-path pairs of explicit imports to use. See [mockgen's -imports](https://github.com/uber-go/mock#flags) for more information.
        copyright_file: optional file containing copyright to prepend to the generated contents. See [mockgen's -copyright_file](https://github.com/uber-go/mock#flags) for more information.
        mock_names: dictionary of interface name to mock name pairs to change the output names of the mock objects. Mock names default to 'Mock' prepended to the name of the interface. See [mockgen's -mock_names](https://github.com/uber-go/mock#flags) for more information.
        kwargs: common attributes](https://bazel.build/reference/be/common-definitions#common-attributes) to all Bazel rules.
    """
    if source:
        _gomock_source(
            name = name,
            out = out,
            library = library,
            source_importpath = source_importpath,
            source = source,
            package = package,
            self_package = self_package,
            aux_files = aux_files,
            mockgen_tool = mockgen_tool,
            mockgen_args = mockgen_args,
            imports = imports,
            copyright_file = copyright_file,
            mock_names = mock_names,
            **kwargs
        )
    else:
        _gomock_archive(
            name = name,
            out = out,
            library = library,
            interfaces = interfaces,
            package = package,
            self_package = self_package,
            mockgen_tool = mockgen_tool,
            copyright_file = copyright_file,
            mock_names = mock_names,
            **kwargs
        )

def _gomock_archive_impl(ctx):
    args = ctx.actions.args()

    if len(ctx.attr.mock_names.items()):
        mock_names = ",".join(["{0}={1}".format(name, pkg) for name, pkg in ctx.attr.mock_names.items()])
        args.add("-mock_names", mock_names)

    args.add("-package", ctx.attr.library[GoInfo].importpath)
    args.add("-archive", ctx.attr.library[GoArchive].data.export_file.path)
    args.add("-destination", ctx.outputs.out)
    args.add("-package", ctx.attr.package)
    args.add("-self_package", ctx.attr.self_package)
    args.add(ctx.attr.library[GoInfo].importpath)
    args.add_joined(ctx.attr.interfaces, join_with = ",")

    # Porting in fix for x/tools >= 0.27.0 from https://github.com/bazel-contrib/rules_go/pull/4173
    go_ctx = go_context(ctx)
    go_ctx.actions.run_shell(
        outputs = [ctx.outputs.out],
        inputs = [ctx.attr.library[GoArchive].data.export_file],
        arguments = [args],
        tools = [ctx.executable.mockgen_tool, go_ctx.sdk.go],
        command = """
            export PATH=$(pwd)/$(dirname {go}):$PATH &&
            export GOROOT=$(pwd)/{goroot} &&
            {cmd} "$@"
        """.format(
            go = go_ctx.go.path,
            goroot = go_ctx.sdk.root_file.dirname,
            cmd = ctx.executable.mockgen_tool.path,
        ),
        mnemonic = "GoMockArchiveGen",
    )

_gomock_archive = rule(
    _gomock_archive_impl,
    attrs = {
        "library": attr.label(
            providers = [GoInfo, GoArchive],
            mandatory = True,
            doc = "The library that contains the interfaces to mock.",
        ),
        "copyright_file": attr.label(
            doc = "Optional file containing copyright to prepend to the generated contents.",
            allow_single_file = True,
            mandatory = False,
        ),
        "out": attr.output(
            mandatory = True,
            doc = "The new Go source file to put the generated mock code.",
        ),
        "interfaces": attr.string_list(
            allow_empty = True,
            doc = "Interfaces to mock.",
        ),
        "package": attr.string(
            doc = "The name of the package the generated mocks should be in. If not specified, uses mockgen's default.",
        ),
        "self_package": attr.string(
            doc = "The full package import path for the generated code.",
        ),
        "mock_names": attr.string_dict(
            doc = "Dictionary of interface name to mock name pairs to change the output names of the mock objects. Mock names default to 'Mock' prepended to the name of the interface.",
            default = {},
        ),
        "mockgen_tool": attr.label(
            default = Label("@com_github_golang_mock//mockgen"),
            executable = True,
            cfg = "exec",
        ),
        "use_underlying_names": attr.bool(
            doc = "Use alias underlying type names in generated mocks instead of the alias names directly",
            default = False,
        ),
    },
    toolchains = [GO_TOOLCHAIN],
)

def _handle_shared_args(ctx, args):
    needed_files = []

    if ctx.attr.package != "":
        args += ["-package", ctx.attr.package]
    if ctx.attr.self_package != "":
        args += ["-self_package", ctx.attr.self_package]
    if len(ctx.attr.imports) > 0:
        imports = ",".join(["{0}={1}".format(name, pkg) for name, pkg in ctx.attr.imports.items()])
        args += ["-imports", imports]
    if ctx.file.copyright_file != None:
        args += ["-copyright_file", ctx.file.copyright_file.path]
        needed_files.append(ctx.file.copyright_file)
    if len(ctx.attr.mock_names) > 0:
        mock_names = ",".join(["{0}={1}".format(name, pkg) for name, pkg in ctx.attr.mock_names.items()])
        args += ["-mock_names", mock_names]
    if ctx.attr.mockgen_args:
        args += ctx.attr.mockgen_args

    return args, needed_files
