package signal

import (
	"math"
	"math/rand"
	"time"
	"unsafe"
)

const (
	Letters         = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	MaximumCapacity = math.MaxInt>>1 + 1
)

var rn = rand.NewSource(time.Now().UnixNano())

// RandSeq generates a random string to serve as dummy data
//
// It returns a deterministic sequence of values each time a program is run.
// Use rand.NewSource() function in your real applications.
func RandSeq(length int) string {
	return random(Letters, length)
}

func random(s string, length int) string {
	// 仿照strings.Builder
	// 创建一个长度为 length 的字节切片
	bytes := make([]byte, length)
	strLength := len(s)
	if strLength <= 0 {
		return ""
	} else if strLength == 1 {
		for i := 0; i < length; i++ {
			bytes[i] = s[0]
		}
		return *(*string)(unsafe.Pointer(&bytes))
	}
	// s的字符需要使用多少个比特位数才能表示完
	// letterIDBits := int(math.Ceil(math.Log2(strLength))),下面比上面的代码快
	letterIDBits := int(math.Log2(float64(nearestPowerOfTwo(strLength))))
	// 最大的字母id掩码
	var letterIDMask int64 = 1<<letterIDBits - 1
	// 可用次数的最大值
	letterIDMax := 63 / letterIDBits
	// 循环生成随机字符串
	for i, cache, remain := length-1, rn.Int63(), letterIDMax; i >= 0; {
		// 检查随机数生成器是否用尽所有随机数
		if remain == 0 {
			cache, remain = rn.Int63(), letterIDMax
		}
		// 从可用字符的字符串中随机选择一个字符
		if idx := int(cache & letterIDMask); idx < strLength {
			bytes[i] = s[idx]
			i--
		}
		// 右移比特位数，为下次选择字符做准备
		cache >>= letterIDBits
		remain--
	}
	// 仿照strings.Builder用unsafe包返回一个字符串，避免拷贝
	// 将字节切片转换为字符串并返回
	return *(*string)(unsafe.Pointer(&bytes))
}

// nearestPowerOfTwo 返回一个大于等于cap的最近的2的整数次幂，参考java8的hashmap的tableSizeFor函数
func nearestPowerOfTwo(cap int) int {
	n := cap - 1
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	if n < 0 {
		return 1
	} else if n >= MaximumCapacity {
		return MaximumCapacity
	}
	return n + 1
}
