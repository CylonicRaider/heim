import React from 'react'
import createReactClass from 'create-react-class'
import PropTypes from 'prop-types'

import Tree from '../Tree'
import ThreadListItem from './ThreadListItem'
import MessageData from '../MessageData'
import TreeNodeMixin from './TreeNodeMixin'


export default createReactClass({
  displayName: 'ThreadList',

  propTypes: {
    tree: PropTypes.instanceOf(Tree).isRequired,
    threadTree: PropTypes.instanceOf(Tree).isRequired,
    threadData: PropTypes.instanceOf(MessageData),
    onScroll: PropTypes.func,
    onThreadSelect: PropTypes.func,
    threadNodeId: PropTypes.string,
  },

  mixins: [
    require('react-immutable-render-mixin'),
    TreeNodeMixin('thread'),
  ],

  getDefaultProps() {
    return {threadNodeId: '__root'}
  },

  render() {
    return (
      <div className="thread-list" onScroll={this.props.onScroll}>
        {this.state.threadNode.get('children').toSeq().map(threadId => <ThreadListItem key={threadId} threadData={this.props.threadData} threadTree={this.props.threadTree} threadNodeId={threadId} tree={this.props.tree} nodeId={threadId} onClick={this.props.onThreadSelect} />)}
      </div>
    )
  },
})
