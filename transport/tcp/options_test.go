package tcp

import (
	"testing"

	"github.com/funkygao/assert"
	"github.com/funkygao/nano"
)

func TestOptionsInvalidName(t *testing.T) {
	opt := newOptions()
	_, err := opt.get(nano.OptionSendDeadline)
	assert.Equal(t, nano.ErrBadOption, err)
	_, err = opt.get(nano.OptionReadQLen)
	assert.Equal(t, nano.ErrBadOption, err)
	_, err = opt.get(nano.OptionTlsConfig)
	assert.Equal(t, nano.ErrBadOption, err)

	err = opt.set(nano.OptionReadQLen, 1)
	assert.Equal(t, nano.ErrBadOption, err)
}

func TestOptionsValidName(t *testing.T) {
	opt := newOptions()
	defaultKeepAlive, err := opt.get(nano.OptionKeepAlive)
	assert.Equal(t, nil, err)
	assert.Equal(t, true, defaultKeepAlive.(bool))
	defaultNoDelay, err := opt.get(nano.OptionNoDelay)
	assert.Equal(t, nil, err)
	assert.Equal(t, true, defaultNoDelay)

	err = opt.set(nano.OptionKeepAlive, false)
	assert.Equal(t, nil, err)
	keepAlive, err := opt.get(nano.OptionKeepAlive)
	assert.Equal(t, nil, err)
	assert.Equal(t, false, keepAlive)

	err = opt.set(nano.OptionNoDelay, false)
	assert.Equal(t, nil, err)
	noDelay, err := opt.get(nano.OptionNoDelay)
	assert.Equal(t, nil, err)
	assert.Equal(t, false, noDelay)
}
