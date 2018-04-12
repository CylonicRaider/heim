import _ from 'lodash'
import url from 'url'
import EventEmitter from 'eventemitter3'


function logPacket(kind, data, highlight) {
  /* eslint-disable no-console */
  const group = highlight ? 'group' : 'groupCollapsed'
  const colors = {
    'send': 'green',
    'recv': '#06f',
    'buffered-send': 'gray',
  }
  console[group](
    '%c%s %c%s %c%s',
    'color: ' + colors[kind], kind,
    'color: black', data.type,
    highlight ? 'background: #efb' : 'color: gray; font-weight: normal', data.id ? '(id: ' + data.id + ')' : '(no id)'
  )
  console.log(data)
  console.log(JSON.stringify(data, true, 2))
  console.groupEnd()
}


export default class Socket {
  constructor() {
    this.endpoint = null
    this.roomName = null
    this.events = new EventEmitter()
    this.ws = null
    this.seq = 0
    this.pingTimeout = null
    this.pingReplyTimeout = null
    this.nextPing = 0
    this.pingLimit = 2000
    this.lastMessage = null
    this._recvBuffer = null
    this._sendBuffer = []
    this._logPackets = false
    this._logPacketIds = {}
  }

  _wsurl(endpoint, roomName) {
    const parsedEndpoint = url.parse(endpoint)

    const prefix = parsedEndpoint.pathname === '/' ? '' : parsedEndpoint.pathname

    let scheme = 'ws'
    if (parsedEndpoint.protocol === 'https:') {
      scheme = 'wss'
    }

    return scheme + '://' + parsedEndpoint.host + prefix + '/room/' + roomName + '/ws?h=1'
  }

  startBuffering() {
    if (!this._recvBuffer) {
      this._recvBuffer = []
    }
  }

  endBuffering() {
    _.each(this._recvBuffer, ([name, data]) =>
      this.events.emit(name, data)
    )
    this._recvBuffer = null
  }

  _emit(name, data) {
    if (this._recvBuffer) {
      this._recvBuffer.push([name, data])
    } else {
      this.events.emit(name, data)
    }
  }

  connect(endpoint, roomName, opts) {
    this._logPackets = opts && opts.log
    this.endpoint = endpoint
    this.roomName = roomName
    this.reconnect()
  }

  reconnect() {
    if (this.ws) {
      // forcefully drop websocket and reconnect
      this._onClose()
      this.ws.close()
    }
    const wsurl = this._wsurl(this.endpoint, this.roomName)
    this.ws = new WebSocket(wsurl, 'heim1')
    this.ws.onopen = this._onOpen.bind(this)
    this.ws.onclose = this._onCloseReconnectSlow.bind(this)
    this.ws.onmessage = this._onMessage.bind(this)
  }

  _onOpen() {
    this._emit('open')
    this._sendBuffer.forEach(item => this._send(item.data, item.log))
    this._sendBuffer = []
  }

  _onClose() {
    clearTimeout(this.pingTimeout)
    clearTimeout(this.pingReplyTimeout)
    this.pingReplyTimeout = null
    this.ws.onopen = this.ws.onclose = this.ws.onmessage = null
    this._emit('close')
  }

  _onCloseReconnectSlow() {
    this._onClose()
    const delay = 2000 + 3000 * Math.random()
    setTimeout(this.reconnect.bind(this), delay)
  }

  _onMessage(ev) {
    const data = JSON.parse(ev.data)

    const packetLogged = _.has(this._logPacketIds, data.id)
    if (this._logPackets || packetLogged) {
      logPacket('recv', data, packetLogged)
    }

    this.lastMessage = Date.now()

    this._handlePings(data)

    this._emit('receive', data)
  }

  _handlePings(msg) {
    if (msg.type === 'ping-event') {
      if (msg.data.next > this.nextPing) {
        const interval = msg.data.next - msg.data.time
        this.nextPing = msg.data.next
        clearTimeout(this.pingTimeout)
        this.pingTimeout = setTimeout(this._ping.bind(this), interval * 1000)
      }

      this.send({
        type: 'ping-reply',
        data: {
          time: msg.data.time,
        },
      })
    }

    // receiving any message removes the need to ping
    clearTimeout(this.pingReplyTimeout)
    this.pingReplyTimeout = null
  }

  _send(data, log) {
    if (!data.id) {
      data.id = String(this.seq++)
    }

    // FIXME: remove when fixed on server
    if (!data.data) {
      data.data = {}
    }

    if (log) {
      this._logPacketIds[data.id] = true
    }
    if (this._logPackets || log) {
      logPacket('send', data, log)
    }
    this.ws.send(JSON.stringify(data))
  }

  send(data, log) {
    if (this.ws.readyState === WebSocket.OPEN) {
      try {
        this._send(data, log)
      } catch (err) {
        // on exception (connection dropped?) reconnect and retry
        console.warn('error sending to socket. reconnecting and retrying.', err, data)  // eslint-disable-line no-console
        this._sendBuffer.push({data, log: !!log})
        this.reconnect()
      }
    } else {
      if (this._logPackets || log) {
        logPacket('buffered-send', data, log)
      }
      this._sendBuffer.push({data, log: !!log})
    }
  }

  _ping() {
    if (this.pingReplyTimeout) {
      return
    }

    this.send({
      type: 'ping',
    })

    this.pingReplyTimeout = setTimeout(this.reconnect.bind(this), this.pingLimit)
  }

  pingIfIdle() {
    if (this.lastMessage === null || Date.now() - this.lastMessage >= this.pingLimit) {
      this._ping()
    }
  }

  on(...args) {
    this.events.on(...args)
  }

  off(...args) {
    this.events.off(...args)
  }

  once(...args) {
    this.events.once(...args)
  }
}
