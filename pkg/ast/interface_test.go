package ast

// Tests the implementations of ItemNode interface.
// The implementations consist of nodes that represent SECS-II data types
// which are ASCII, binary, boolean, float(4,8), int(1,2,4,8), uint(1,2,4,8).
//
// Testing Strategy:
//
// For each implementation, create a new instance using the factory method or FillVariables(),
// and test the result of public observer methods Size(), Variables(), ToBytes(), and String().
//
// Partitions:
//
// - The size of the node (The number of data values it include): 0, 1, ...
// - The number of variables in the node: 0, 1, ...
// - Partitions of the data values for each node:
//   - Boolean data: true, false
//   - Binary data: 0, 1, ..., 255 (for integer and binary string representation)
//   - Float data: Min, ..., -SmallestNonZero, 0, SmallestNonZero, ..., Max
//   - Int data: Min, ..., -1, 0, 1, ..., Max
//   - Uint data: 0, 1, ..., Max
//
// * ASCII and List type are special types and the partitions might differ.
//   Refer to ascii_test.go and list_test.go.
