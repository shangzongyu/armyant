// Copyright 2014 hey Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package work

import (
	"fmt"
	"net/url"
	"strings"
	"sync"
	_ "time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type MqttWork struct {
	client       MQTT.Client
	waitingQueue map[string]func(client MQTT.Client, msg MQTT.Message)
	lock         *sync.Mutex
	currId       int64
}

func (mw *MqttWork) GetDefaultOptions(addrURI string) *MQTT.ClientOptions {
	mw.currId = 0
	mw.lock = new(sync.Mutex)
	mw.waitingQueue = make(map[string]func(client MQTT.Client, msg MQTT.Message))
	opts := MQTT.NewClientOptions()
	opts.AddBroker(addrURI)
	opts.SetClientID("1")
	opts.SetUsername("")
	opts.SetPassword("")
	opts.SetCleanSession(false)
	opts.SetProtocolVersion(3)
	opts.SetAutoReconnect(false)
	opts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
		//收到消息
		mw.lock.Lock()
		if callback, ok := mw.waitingQueue[msg.Topic()]; ok {
			//有等待消息的callback 还缺一个信息超时的处理机制
			_, err := url.Parse(msg.Topic())
			if err != nil {
				ts := strings.Split(msg.Topic(), "/")
				if len(ts) > 2 {
					//这个topic存在msgid 那么这个回调只使用一次
					delete(mw.waitingQueue, msg.Topic())
				}
			}
			go callback(client, msg)
		}
		mw.lock.Unlock()
	})
	return opts
}
func (mw *MqttWork) Connect(opts *MQTT.ClientOptions) error {
	//fmt.Println("Connect...")
	mw.client = MQTT.NewClient(opts)
	if token := mw.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (mw *MqttWork) GetClient() MQTT.Client {
	return mw.client
}

func (mw *MqttWork) Finish() {
	mw.client.Disconnect(250)
}

/**
 * 向服务器发送一条消息
 * @param topic
 * @param msg
 * @param callback
 */
func (mw *MqttWork) Request(topic string, body []byte) (MQTT.Message, error) {
	mw.currId = mw.currId + 1
	topic = fmt.Sprintf("%s/%d", topic, mw.currId) //给topic加一个msgid 这样服务器就会返回这次请求的结果,否则服务器不会返回结果
	result := make(chan MQTT.Message)
	mw.On(topic, func(client MQTT.Client, msg MQTT.Message) {
		result <- msg
	})
	mw.GetClient().Publish(topic, 0, false, body)
	msg, ok := <-result
	if !ok {
		return nil, fmt.Errorf("client closed")
	}
	return msg, nil
}

/**
 * 向服务器发送一条消息
 * @param topic
 * @param msg
 * @param callback
 */
func (mw *MqttWork) RequestURI(u *url.URL, body []byte) (MQTT.Message, error) {
	mw.currId = mw.currId + 1
	v := u.Query()
	v.Add("msg_id", fmt.Sprintf("%v", mw.currId)) //给topic加一个msgid 这样服务器就会返回这次请求的结果,否则服务器不会返回结果
	u.RawQuery = v.Encode()
	topic := u.String()
	result := make(chan MQTT.Message)
	mw.On(topic, func(client MQTT.Client, msg MQTT.Message) {
		result <- msg
	})
	mw.GetClient().Publish(topic, 0, false, body)
	msg, ok := <-result
	if !ok {
		return nil, fmt.Errorf("client closed")
	}
	return msg, nil
}

/**
 * 向服务器发送一条消息
 * @param topic
 * @param msg
 * @param callback
 */
func (mw *MqttWork) RequestURINR(url *url.URL, body []byte) {
	mw.currId = mw.currId + 1
	v := url.Query()
	url.RawQuery = v.Encode()
	topic := url.String()
	mw.GetClient().Publish(topic, 0, false, body)
}

/**
 * 向服务器发送一条消息,但不要求服务器返回结果
 * @param topic
 * @param msg
 */
func (mw *MqttWork) RequestNR(topic string, body []byte) {
	mw.GetClient().Publish(topic, 0, false, body)
}

/**
 * 监听指定类型的topic消息
 * @param topic
 * @param callback
 */
func (mw *MqttWork) On(topic string, callback func(client MQTT.Client, msg MQTT.Message)) {
	////服务器不会返回结果
	mw.lock.Lock()
	mw.waitingQueue[topic] = callback //添加这条消息到等待队列
	mw.lock.Unlock()
}
