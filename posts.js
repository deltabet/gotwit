
var Post = React.createClass({
	render: function(){
		return (
			<div className="post">
				{this.props.author}</br>
			</div>
		)
	}
});


var PostList = React.createClass({
	render: function(){
		var postNodes = this.props.data.map(function(post) {
			return(
				<Post author={post.author} key={post.id}>
					{post.text}
				</Post>
			);
		});
		return(
			<div className="postList">
				{postNodes}
			</div>
		);
	}
});

var PostBox = React.createClass({
	loadPosts: function() {
    $.ajax({
      url: this.props.url,
      dataType: 'json',
      cache: false,
      success: function(data) {
        this.setState({data: data});
      }.bind(this),
      error: function(xhr, status, err) {
        console.error(this.props.url, status, err.toString());
      }.bind(this)
    });
  },
	handlePostSubmit: function(post){
		var posts = this.state.data;
		//id
		var newPosts = posts.concat([post]);
		this.setState({data: newPosts});
		$.ajax({
			url: this.props.url,
			dataType: 'json',
			type: 'POST',
			data: post,
			success: function(data){
				this.setState({data: data});
			}.bind(this),
			error: function(xhr, status, err){
				this.setState({data: posts});
				console.error(this.props.url, status, err.toString());
			}.bind(this)
		});
	},
	getInitialState: function(){
		return {data: []};
	},
	componentDidMount: function() {
    this.loadPosts();
    setInterval(this.loadPosts, this.props.pollInterval);
  },
	render: function(){
		return(
			<div className="postBox">
				<PostList data={this.state.data} />
				<PostForm onPostSubmit = {this.handlePostSubmit} />
			</div>
		)
	}
});

var PostForm = React.createClass({
	getInitialState: function(){
		return {author: '', text: ''};
	},
	handleAuthorChange: function(e) {
    this.setState({author: e.target.value});
  },
  handleTextChange: function(e) {
    this.setState({text: e.target.value});
  },
	handleSubmit: function(e){
		e.preventDefault();
		var author = this.state.author;
		var text = this.state.text;
		if (!text || !author){
			return;
		}
		this.props.onPostSubmit({author: author, text: text});
		this.setState({author: '', text: ''});
	},
	render: function(){
		return(
			<form className="commentForm" onSubmit={this.handleSubmit}>
				<input
					type="text"
					value = {this.state.author}
					onChange={this.handleAuthorChange}
				/>
				<input
					type="text"
					value={this.state.text}
					onChange={this.handleTextChange}
				/>
				<input type="submit" value="POST" />
			</form>	
		);
	}
});

ReactDOM.render(
  <PostBox url="/page/new" pollInterval={2000}/>,
  document.getElementById('posts')
);
