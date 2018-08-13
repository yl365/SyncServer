package main

import (
	"time"

	"github.com/gomodule/redigo/redis"
)

func newPool(addr, password string, db int) *redis.Pool {
	return &redis.Pool{

		MaxActive:   900,
		MaxIdle:     30,
		Wait:        true,
		IdleTimeout: 300 * time.Second,

		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < 10*time.Second {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},

		Dial: func() (redis.Conn, error) {

			c, err := redis.DialTimeout("tcp", addr, 3*time.Second, 1*time.Second, 1*time.Second)
			if err != nil {
				return nil, err
			}

			if len(password) > 0 {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}

			if _, err := c.Do("SELECT", db); err != nil {
				c.Close()
				return nil, err
			}
			return c, nil
		},
	}
}
