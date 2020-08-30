package version

import "log"

type Compaction struct {
	level  int
	inputs [2][]*FileMetaData
}

func (c *Compaction) isTrivialMove() bool {
	return len(c.inputs[0]) == 1 && len(c.inputs[1]) == 0
}

func (c *Compaction) Log() {
	log.Printf("Compaction, level: %d", c.level)
	for i := 0; i < len(c.inputs[0]); i++ {
		log.Printf("inputs[0]: %d", c.inputs[0][i].number)
	}

	for i := 0; i < len(c.inputs[1]); i++ {
		log.Printf("inputs[1]: %d", c.inputs[1][i].number)
	}
}

func totalFileSize(files []*FileMetaData) uint64 {
	var sum uint64
	for i := 0; i < len(files); i++ {
		sum += files[i].fileSize
	}
	return sum
}

func maxBytesForLevel(level int) float64 {
	result := 10. * 1048576.0
	for level > 1 {
		result *= 10
		level--
	}
	return result
}
