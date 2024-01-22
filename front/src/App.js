import React, { useState } from 'react';
import axios from 'axios';
import './App.css';

const App = () => {
  const [formData, setFormData] = useState({
    plan: '',
    companyName: '',
    iban: '',
    bic: '',
    address: ''
  });
  const [response, setResponse] = useState({
    link: '',
    message: '',
    error: {
      message: '',
      fields: ''
    }
  });

  const handleInputChange = (e) => {
    const { name, value } = e.target;
    setFormData(prevState => ({
      ...prevState,
      [name]: value
    }));
  };

  const handleSubmit = async (event) => {
    event.preventDefault();
    console.log(JSON.stringify(formData));

    // Use FormData to construct the data payload
    const data = new FormData();
    data.append('plan', formData.plan);
    data.append('companyName', formData.companyName);
    data.append('iban', formData.iban);
    data.append('bic', formData.bic);
    data.append('address', formData.address);


    axios.post('http://localhost:8080/create', data, {
      headers: {
        Authorization: 'hds8fg8dhfs&g7sd9^' // Replace with actual token
      }
      })
      .then(res => {
        console.log(res.data);
        //TODO: handle response AND FIX ERROR
        setResponse(res.data);
      })
      .catch(err => {
        // handle error
        console.log(err.response.data);
        setResponse({ message: err.response.data.message + ': ' + err.response.data.fields });
      });
  };

  return (
    <div style={{ display: 'flex', justifyContent: 'space-between' }}>
      <form style={{ width: '50%' }} onSubmit={handleSubmit}>
        <input
          type="text"
          name="plan"
          placeholder="Plan"
          value={formData.plan}
          onChange={handleInputChange}
        />
        <input
          type="text"
          name="companyName"
          placeholder="Company Name"
          value={formData.companyName}
          onChange={handleInputChange}
        />
        <input
          type="text"
          name="iban"
          placeholder="IBAN"
          value={formData.iban}
          onChange={handleInputChange}
        />
        <input
          type="text"
          name="bic"
          placeholder="BIC"
          value={formData.bic}
          onChange={handleInputChange}
        />
        <input
          type="text"
          name="address"
          placeholder="Address"
          value={formData.address}
          onChange={handleInputChange}
        />
        <button type="submit">Create</button>
      </form>

      <div style={{ width: '50%' }}>
        <p>Plan: {formData.plan}</p>
        <p>Company Name: {formData.companyName}</p>
        <p>IBAN: {formData.iban}</p>
        <p>BIC: {formData.bic}</p>
        <p>Address: {formData.address}</p>
      </div>

      <div style={{ width: '100%' }}>
        <p>Response: {response.message}</p>
        <p>{response.link ? <a target='_blank' href={'http://' + response.link}>Download pass</a> : ''}</p>
      </div>
    </div>
  );
};

export default App;
