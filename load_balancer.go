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



// The load balancer uses a priority queue to divide up subtree searches evenly among available 'worker'
// goroutines, based on the expected effort required to search the subtree and its relative importance
// based on how promising that part of the tree appears to be.


// Expected effort is the average branching factor for the game tree raised by the depth remaining to search.


// Subtrees rooted along the Principal Variation (PV-Nodes) should be searched first, 
// followed by fail-high nodes (Cut-Nodes). Fail-low nodes (All-Nodes) should be searched last.
// Nodes of the same type should be processed left to right (making use of move ordering heuristics)




























