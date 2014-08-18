//-----------------------------------------------------------------------------------
// Copyright (c) 2014 Stephen J. Lovell
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
//-----------------------------------------------------------------------------------

package load_balancer


type Worker struct {
  requests chan Request // work to do (a buffered channel)
  pending  int          // count of pending tasks
  index    int          // index in the heap
}

type Done struct {
  w *Worker
  size int
}

func (w *Worker) work(done chan Done) {
  go func() {
    for {
      select {
      case req := <-w.requests: // get requests from load balancer
        go func() {
          req.Result <- req.Fn() // do the work and send the answer back to the requestor
          // println("Result sent")
          done <- Done{w, req.Size}         // tell load balancer a task has been completed by worker w.
          // println("Done sent")
        }()
      }
    }
  }()

}

// work() will block forever waiting for result

// Spawning a new goroutine with each executed request defeats the purpose of the load balancer.

// To avoid this, would need to come up with some way of managing incoming requests so that workers
// are never blocked waiting on a request further up in its own queue.

// One option would be to separate requests by ply, 
// with a separate load balancer / worker pool for incoming requests from each ply.


















