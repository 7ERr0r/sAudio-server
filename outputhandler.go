package sAudio-server



type OutputHandler struct {
	sess *AudioClient
	inputc chan []float32
	
}
func NewOutputHandler(sess *AudioClient)(this *OutputHandler){
	this = new(OutputHandler)
	this.sess = sess
	this.inputc = make(chan []float32, 64)
	this.sess.Log("output handler")

	return
}
func (this *OutputHandler) Receive(data []byte)(err error){
	return
}
func (this *OutputHandler) HandleBroadcast(data []float32){
	select {
	case this.inputc <- data:
	default:
	}
}
func (this *OutputHandler) Serve()(err error){
	enc := new(opus.Encoder)
	err = enc.Init(48000, 2, opus.APPLICATION_AUDIO)
	if err != nil {
		err = fmt.Errorf("could not create opus encoder: %s", err.Error())
		return
	}


	buf := make([]byte, 4000)
	var n int
	var sequenceId uint32
	for rawdata := range this.inputc {
		binary.LittleEndian.PutUint32(buf, sequenceId)
		n, err = enc.EncodeFloat32(rawdata, buf[4:])
		if err != nil {
			return
		}
		//if rand.Float32() > 0.05 { // Packet loss simulator

		err = this.sess.Send(buf[:n+4])
		if err != nil {
			return
		}

		//}
		sequenceId++
	}

	return
}