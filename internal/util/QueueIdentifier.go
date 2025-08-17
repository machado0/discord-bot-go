package util

type QueueIdentifier struct {
	idToName map[int]string
}

func NewQueueIdentifier() *QueueIdentifier {
	return &QueueIdentifier{
		idToName: map[int]string{
			420: "Solo Queue",
			400: "Normal Game",
			440: "Flex",
			450: "ARAM",
		},
	}
}

func (qi *QueueIdentifier) GetQueueNameByID(id int) string {
	if name, ok := qi.idToName[id]; ok {
		return name
	}
	return "Não identificado"
}