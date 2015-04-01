package main

type mailID string

type mailIDSlice []mailID

func (p mailIDSlice) Len() int {
    return len(p)
}

func (p mailIDSlice) Less(i, j int) bool {
    return p[i] < p[j]
}

func (p mailIDSlice) Swap(i, j int) {
    p[i], p[j] = p[j], p[i]
}

