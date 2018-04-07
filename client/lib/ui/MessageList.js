import React from 'react'
import createReactClass from 'create-react-class'
import PropTypes from 'prop-types'

import Message from './Message'
import Tree from '../Tree'
import { Pane } from '../stores/ui'
import TreeNodeMixin from './TreeNodeMixin'


export default createReactClass({
  displayName: 'MessageList',

  propTypes: {
    pane: PropTypes.instanceOf(Pane).isRequired,
    tree: PropTypes.instanceOf(Tree).isRequired,
    showTimeStamps: PropTypes.bool,
    roomSettings: PropTypes.object,
    nodeId: PropTypes.string,
    depth: PropTypes.number,
  },

  mixins: [
    require('react-immutable-render-mixin'),
    TreeNodeMixin(),
  ],

  getDefaultProps() {
    return {nodeId: '__root', depth: 0}
  },

  componentDidMount() {
    this.props.pane.messageRenderFinished()
  },

  componentDidUpdate() {
    this.props.pane.messageRenderFinished()
  },

  render() {
    const children = this.state.node.get('children')
    return (
      <div className="message-list">
        {children.toIndexedSeq().map((nodeId, idx) =>
          <Message key={nodeId} pane={this.props.pane} tree={this.props.tree} nodeId={nodeId} showTimeAgo={idx === children.size - 1} showTimeStamps={this.props.showTimeStamps} roomSettings={this.props.roomSettings} />
        )}
      </div>
    )
  },
})
