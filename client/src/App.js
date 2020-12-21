import './App.css';
import React from 'react';
import Image from 'react-bootstrap/Image'
import axios from 'axios';
import { Fragment } from 'react';
const resultStart = '<p>&nbsp;&nbsp;&nbsp;&nbsp;<span class="ru">Result:'
const resultEnd = '......."</p>'


const config = {
  timeout: 180000,
  headers: { 'Content-Type': 'application/json', 'Accept': 'text/xml' }
};
class App extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      value: '',
      results: '',
      error: ''
    }
    this.onChange = this.onChange.bind(this);
    this.onSearch = this.onSearch.bind(this);
    this.onFocus = this.onFocus.bind(this);
    this.onBlur = this.onBlur.bind(this);
  }
  onChange(event) {
    this.setState({ value: event.target.value });
  }
  onFocus(event) {
    event.target.placeholder = '';
    this.setState({ value: event.target.value });
  }
  onBlur(event) {
    event.target.placeholder = '';
    this.setState({ value: event.target.value });
  }
  async validateForm (event){
    this.setState({ value: this.state.value });
    if (this.state.value.length >= 1) {
      this.state.error = '';
      return true
    } else {
      this.state.error = 'Please enter something search.';
      return false
    }
  }
  async onSearch(event){
    let valid = await this.validateForm(event);
    if (valid) {
      event.preventDefault();
      axios.get('/search?q=' + this.state.value, config)
        .then((res) =>{
         let out = ''
         let n = 0;          
          if (res.data.length) {
             for (let i = 0; i < res.data.length; i++) {
              n++;                         
              let text = resultStart + n + '</span>&nbsp;&nbsp;".......' + res.data[i] +  resultEnd;
              out = out += text
            }
            this.setState({ results: out });
          } else {
            this.setState({ results: 'No search results found.' });
          }
        }).catch ((error) =>{
          if (error) {
            this.setState({ results: 'Search error please try search again' });
            console.log('Error', error.message);
          }
        });  
    
  }
}
  render() {
    const { results, value, error } = this.state;
    return <Fragment>
      <div className="p-2 flex-md-column align-middle">
        <div className="input-group mb-3">
          <button className="sbtn-primary input-group-btn shadow-none" onClick={this.onSearch}><img src="Vector.png" /><span className="fnt">Search</span></button>&nbsp;&nbsp;&nbsp;&nbsp;
         <input className="spad shadow-none sf form-control" type="text" value={value} onChange={this.onChange} onFocus={this.onFocus} onBlur={this.onBlur} id="search" placeholder="What art thee looking f'r?" />
        </div>
        <div className="input-group sm-3">
        <p className="text-danger">{error}</p>         
        </div>
      </div>
      <Result results={results} />
    </Fragment>
  }
}
const Result = ({ results }) => {
  if (results) {
    return <div className="p-2 flex-sm-column text2 tf" dangerouslySetInnerHTML={{ __html: results }} />
  } else {
    return <div className="p-2 flex-sm-column align-right"> <Image className="im" alt="" src="image28.png"></Image></div>;
  }
}
export default App;
