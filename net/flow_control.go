package util

import (
	"sync"
	"time"
)

type FlowControl struct {
	sync.Mutex
	grid_count   uint64   //格子个数
	grid_distant uint64   //格子距离
	grid_array   []uint64 //每个格子的存量

	start_time time.Time //启动时间
	max_count  uint64    //最大存量
	cur_count  uint64    //当前存量

	last_grid         uint64 //上次的实际格子
	virtual_last_grid uint64 //上次的虚拟格子
}

//参数：格子个数，每个格子的时长(ms),总容量 。在 grid_count * grid_distant 毫秒内最多通过 max_count请求
func NewFlowControl(grid_count uint64, grid_distant uint64, max_count uint64) *FlowControl {
	return &FlowControl{
		grid_count:   grid_count,
		grid_distant: grid_distant,
		grid_array:   make([]uint64, grid_count),
		max_count:    max_count,
		last_grid:    0,
		cur_count:    0,
		start_time:   time.Now(),
	}
}

//当前格子中的容量
func (this *FlowControl) GetCount() uint64 {
	return this.cur_count
}

// 0 : 正常 1:过载
func (this *FlowControl) CheckLoad() int {
	this.Lock()
	defer this.Unlock()
	cur_time := time.Now()
	time_used := uint64(cur_time.Sub(this.start_time).Nanoseconds() / 1000)
	if time_used < 0 {
		this.cur_count = 0
		this.start_time = cur_time
		this.last_grid = 0
		this.virtual_last_grid = 0
		this.grid_array = make([]uint64, this.grid_count)
		return 0
	}

	virtual_curr_grid := time_used / uint64(1000) / this.grid_distant
	curr_grid := virtual_curr_grid % this.grid_count

	//清除过期格子里的存量
	if virtual_curr_grid != this.virtual_last_grid {
		grid_spand := virtual_curr_grid - this.virtual_last_grid
		this.virtual_last_grid = virtual_curr_grid
		if grid_spand > this.grid_count {
			grid_spand = this.grid_count
		}

		var i uint64 = 0
		for i = 0; i < grid_spand; i++ {
			this.cur_count -= this.grid_array[(curr_grid-i+this.grid_count)%this.grid_count]
			this.grid_array[(curr_grid-i+this.grid_count)%this.grid_count] = 0
		}
	}

	//过载
	if this.cur_count+1 >= this.max_count {
		return 1
	}
	this.grid_array[curr_grid]++
	this.cur_count++

	return 0
}
