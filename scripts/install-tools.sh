#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
set -o errtrace
(shopt -p inherit_errexit &>/dev/null) && shopt -s inherit_errexit

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
source "$SCRIPT_DIR/init.bash" || exit 1

l.trap_error "$@"

st.doing "Installing direnv"
st.do go install github.com/direnv/direnv/v2@latest
st.done

st.doing "Installing buf..."
st.do go install github.com/bufbuild/buf/cmd/buf@latest
st.done

st.doing "Installing protoc-gen-go..."
st.do go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
st.done

st.doing "Installing protoc-gen-connect-go..."
st.do go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
st.done

# st.doing "Installing migrate..."
# st.do go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

st.done "Tools installed successfully!"

[ -z "$APP_ENV" ] && {
    echo 'APP_ENV not set. Process aborted'
    exit 1
}

if [ "$APP_ENV" = "development" ]; then
    st.h1 "Installing development tools..."

    st.doing "Installing Goda"
    st.do go install github.com/loov/goda@latest
    st.done

    st.doing "Installing gopls (LSP)"
    st.do go install golang.org/x/tools/gopls@latest
    st.done

    st.doing "Installing golangci-lint LSP wrapper"
    st.do go install github.com/nametake/golangci-lint-langserver@latest
    st.done

    st.doing "Installing  Formatting"
    st.do go install golang.org/x/tools/cmd/goimports@latest
    st.done

    st.doing "Installing  Debugger"
    st.do go install github.com/go-delve/delve/cmd/dlv@latest
    st.done

    st.doing "Installing Static Analysis"
    st.do go install honnef.co/go/tools/cmd/staticcheck@latest
    st.done

    st.doing "Installing Protobuf / gRPC generators"
    st.do go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    st.do go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    st.done

    st.doing "Installing Shell Formatting"
    st.do go install mvdan.cc/sh/v3/cmd/shfmt@latest
    st.done

    st.doing "Installing godef"
    st.do go install github.com/rogpeppe/godef@latest
    st.done

    # Test runner (it is in the go tool)
    # go install gotest.tools/gotestsum@latest

    # Go enum generator (It's in the go tool)
    # st.do go install github.com/abice/go-enum

    st.h1 "Installing dep-tree..."
    if ! command -v dep-tree >/dev/null 2>&1; then
        if command -v brew >/dev/null 2>&1; then
            st.doing "Using brew to install dep-tree..."
            st.do brew install dep-tree
            st.done
        elif command -v pip >/dev/null 2>&1; then
            st.doing "Using pip to install dep-tree..."
            st.do pip install dep-tree
            st.done
        elif command -v npm >/dev/null 2>&1; then
            st.doing "Using npm to install dep-tree..."
            st.do npm install -g dep-tree
            st.done
        else
            st.warn "Warning: Could not install dep-tree - no package manager found (brew, pip, or npm)"
        fi
    else
        st.nothingTodo
    fi

    st.done "Development tools installed successfully!"
fi
