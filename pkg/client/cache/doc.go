// Package cache is a client-side caching mecanism. It is useful for
// reducing the number of server calls you'd otherwise need to make.
// Reflector watches a serer and updates a Store. Two stores are provided;
// one that simply caches objects (for example, to allow a scheduler to
// list currently available minions), and one that additionally acts as
// a FIFO queue (for example, to allow a scheduler to process incoming
// pods).
package cache
