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

package main



// At each node, call each new subtree search in a new goroutine.

// values are piped up the subtrees.  
// When this causes bounds in the node to update, the updated bounds are piped down the subtrees.



// Young Brothers Wait (YBW) approach

// at each node, search the leftmost child sequentially before searching the rest of the successors concurrently.


// When spawning goroutines, priority should be based on node type of subtree root.  
// If the node type is the same, use the move ordering to guess priority.  
// If all goroutines are spawned into a single pool, this would create a "tree splitting" effect.


// Goal is to avoid wasted processing effort where a subtree is expanded that otherwise would have been pruned.



// The more edges have been already explored, the more likely it is that all moves will need to be searched. Could increment a max number
// of goroutines for the current subtree root as the number of moves explored increases. 





// Generate moves in batches to save effort on move generation when cutoffs occur.
// PV, hash, promotions, winning captures, killers, losing captures, quiet moves




// Load balancing

// ???











