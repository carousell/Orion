//Package wrap provides multiple wrap functions to wrap log implementation of other log packages
package wrap

/*
 * gokitwrap wraps the gokit.Logger impl with log.Logger
 */
/*
type gokitwrap struct {
	l log.Logger
}

func (g *gokitwrap) Log(keyvals ...interface{}) error {
	vals := make([]interface{}, 0)
	ctx := context.Background()
	for _, val := range keyvals {
		if c, ok := val.(context.Context); ok {
			ctx = c
		} else {
			vals = append(vals, val)
		}
	}
	g.l.Log(ctx, loggers.InfoLevel, 1, vals...)
	return nil
}

// ToGoKitLogger wraps a log.Logger to gokit/log.Logger
func ToGoKitLogger(l log.Logger) basegokit.Logger {
	return &gokitwrap{
		l: l,
	}
}
*/
