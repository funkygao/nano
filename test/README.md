test
====

The reason test cases has a dedicated dir is to avoid the problem of
golang fordiding circular import.

For example, `util_test.go`, if has no dedicated test pkg, will lead to 
import circle:

    util_test.go import github.com/funkygao/nano/transport/tcp
    transport/tcp will import github.com/funkygao/nano

    thus, lead to circular import!
