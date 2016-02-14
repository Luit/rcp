// Copyright © 2016 Luit van Drongelen <luit@luit.eu>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// rcp connects regular Redis-using applications to a Redis (3.0+) Cluster
package main // import "luit.eu/rcp"

import "luit.eu/rcp/cmd"

func main() {
	cmd.Execute()
}

// Step 1: Dumb mode
//  Just connect to the cluster through the known host and try to use it.
//  Every time a reply comes back as -MOVED or -ASK, disconnect from the
//  current node, and connect to whatever the -MOVED or -ASK pointed to.
//
// Step 2: Slightly smarter
//  Inspect commands to see which node to route to, keeping a slot state of
//  the cluster to ask the right node most of the queries. Still without
//  pooling.
//
// Step 3:
//  Pooling (configurable), pipelining (configurable), marking commands as
//  safe for multiplexing backend connections.
