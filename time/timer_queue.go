package time

import (
	"container/list"
	"sync"
	"time"
)

//定时队列 超时的自动删除
type TimerQueue struct {
	timeout       time.Duration
	checkInterval time.Duration
	exitChan      chan struct{}
	itemList      *list.List
	itemMap       map[uint32]*list.Element
	itemLock      sync.Mutex
}

//队列元素
type Object interface {
	Expire()
}

type timerQueueItem struct {
	EnObject Object
	EnTime   time.Time
	EnSeq    uint32
}

func NewTimerQueue(timeout time.Duration, checkInterval time.Duration) *TimerQueue {
	return &TimerQueue{
		timeout:       timeout,
		checkInterval: checkInterval,
		exitChan:      make(chan struct{}),
		itemList:      list.New(),
		itemMap:       make(map[uint32]*list.Element),
	}
}

func (this *TimerQueue) expireLoop() {
	tick := time.NewTicker(this.checkInterval)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			for {
				this.itemLock.Lock()

				delItem := this.itemList.Front()
				if delItem == nil {
					this.itemLock.Unlock()
					break
				}
				item := delItem.Value.(*timerQueueItem)
				if time.Now().Sub(item.EnTime) < this.timeout {
					this.itemLock.Unlock()
					break
				}
				item.EnObject.Expire()
				delete(this.itemMap, item.EnSeq)
				this.itemList.Remove(delItem)
				this.itemLock.Unlock()
			}
		case <-this.exitChan:
			goto exit
		}
	}
exit:
}

func (this *TimerQueue) Stop() {
	close(this.exitChan)
}

func (this *TimerQueue) Start() {
	go this.expireLoop()
}

func (this *TimerQueue) Size() int {
	return len(this.itemMap)
}

func (this *TimerQueue) IsExist(seq uint32) bool {
	this.itemLock.Lock()
	defer this.itemLock.Unlock()
	_, exist := this.itemMap[seq]
	return exist
}

func (this *TimerQueue) EnQueue(seq uint32, item Object) {
	item_queue := &timerQueueItem{
		EnTime:   time.Now(),
		EnSeq:    seq,
		EnObject: item,
	}

	this.itemLock.Lock()
	defer this.itemLock.Unlock()
	exist_item, exist := this.itemMap[seq]
	if exist {
		delete(this.itemMap, seq)
		this.itemList.Remove(exist_item)
	}

	element := this.itemList.PushBack(item_queue)
	this.itemMap[seq] = element
	return
}

func (this *TimerQueue) DeQueue(seq uint32) Object {
	this.itemLock.Lock()
	defer this.itemLock.Unlock()

	exist_item, exist := this.itemMap[seq]
	if !exist {
		return nil
	}
	item := exist_item.Value.(*timerQueueItem)
	delete(this.itemMap, seq)
	this.itemList.Remove(exist_item)
	return item.EnObject
}

