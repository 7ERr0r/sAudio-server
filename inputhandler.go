package sAudio-server




type InputHandler struct {
	sess *AudioClient
	outputc chan []float32
	dec *opus.Decoder
}
func NewInputHandler(sess *AudioClient)(this *InputHandler){
	this = new(InputHandler)
	this.sess = sess
	this.sess.Log("input handler")

	this.outputc = make(chan []float32, 64)
	
	var err error
	this.dec, err = opus.NewDecoder(SAMPLE_RATE, CHANNELS)
	if err != nil || this.dec == nil {
		this.sess.Logf("err creating opus decoder: %s", err.Error())
		return
	}
	return
}
func (this *InputHandler) Serve()(err error){
	
	return
}
func (this *InputHandler) Receive(data []byte)(err error){
	if len(data) <= 4 {
		return
	}
	sequenceId := binary.LittleEndian.Uint32(data)
	_ = sequenceId

	const FRAME_SIZE_MS = 5
	const FRAME_SIZE = (CHANNELS * SAMPLE_RATE * FRAME_SIZE_MS) / 1000
	rawdata := make([]float32, FRAME_SIZE)
	_, err = this.dec.DecodeFloat32(data[4:], rawdata)
	if err != nil {
		return
	}
	//this.sess.Logf("n: %d, fs: %d", n, FRAME_SIZE)
	//this.outputc <- rawdata[:n]
	this.sess.server.WriteAudio(rawdata)
	return
}