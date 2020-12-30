package worker

import (
	"fmt"
	agentModel "github.com/headend/share-module/model/agentd"
	socketio_client "github.com/zhouhui8915/go-socket.io-client"
	"log"
	"time"
)

type Pool struct {
	Tasks []*Task
	Concurrency int
	TasksChan   chan *Task
	SocConn 	*socketio_client.Client
}


func NewPool(tasks []*Task, concurrency int) *Pool {
	return &Pool{
		Tasks:       tasks,
		Concurrency: concurrency,
		TasksChan:   make(chan *Task),
	}
}


func (p *Pool) Run() {
	//log.Println("Spawn ", worker_num, " worker(s)")
	for i := 0; i <= p.Concurrency; i++ {
		go p.work(i)
	}
}

func (p *Pool) work(worker_id int) {
	for task := range p.TasksChan {
		task.Run(worker_id, p.SocConn)
	}
}

//--------------------------------------------------------------

type Task struct {
	Err 			error
	AgentID			int64
	MonitorID		int64
	ProfileID		int64
	MulticastIP		string
	Status			int64
	SignalStatus	bool
	VideoStatus		bool
	AudioStatus		bool
}


func NewTask(profileMonitor agentModel.ProfileForAgentdElement) *Task {
	return &Task{
		Err:          nil,
		AgentID:      profileMonitor.AgentId,
		MonitorID:    profileMonitor.MonitorId,
		ProfileID:    profileMonitor.ProfileId,
		MulticastIP:  profileMonitor.MulticastIP,
		Status:       profileMonitor.Status,
		VideoStatus:  profileMonitor.VideoStatus,
		AudioStatus:  profileMonitor.AudioStatus,
	}
}


func (t *Task) Run(worker_id int, SocConn *socketio_client.Client) {
	fmt.Println("worker", worker_id, "processing profileip ", t.MulticastIP)
	t.Err = t.CheckSource(SocConn)
}

func (t *Task)CheckSource(SocConn *socketio_client.Client) (err error) {
	log.Println("check source")
	time.Sleep(10*time.Second)
	log.Println("check source done")
	//err, checkcode := selfutils.CheckSourceMulticast(t.MulticastIP)
	//if checkcode != t.Status {
	//	time.Sleep(60*time.Second)
	//	//recheck
	//	err, checkcode = selfutils.CheckSourceMulticast(t.MulticastIP)
	//	if checkcode != t.Status {
	//		msg := fmt.Sprintf("Status has change from %d to %d\n", t.Status, checkcode)
	//		err2 := SocConn.Emit("monitor-response", msg)
	//		if err2 != nil {
	//			log.Println(err2)
	//			return err2
	//		}
	//	}
	//}
	SocConn.Emit("monitor-response", "done")
	return err
}

