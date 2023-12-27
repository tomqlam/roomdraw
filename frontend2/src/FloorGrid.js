import logo from './logo.svg';
import React from 'react';
import { useState } from 'react';
import 'bulma/css/bulma.min.css';

function FloorGrid({cellColors, gridData, dropdownOptions, updateGridData}) {
  // Define your data structure with columns



  const [isOpen, setIsOpen] = useState(false);
  const [selectedItem, setSelectedItem] = useState(null);
  const [dropdownValues, setDropdownValues] = useState(['', '', '']);
  const [pullMethod, setPullMethod] = useState('');
  const [showModalError, setShowModalError] = useState(false);
  


  

  const gridContainerStyle = {
    display: 'grid',
    gridTemplateColumns: 'auto auto auto auto auto', // Adjust the number of 'auto' as needed
    gap: '5px',
    maxWidth: '800px', // Set the maximum width of the grid container
    margin: '0 auto', // Center the grid container horizontally
  };

  const gridItemStyle = {
    //border: '1px solid #ddd',
    borderRadius: '3px',
    padding: '2px',
    backgroundColor: cellColors.occupants, // Set the background color of the cells
    textAlign: 'center',
  };
  const roomNumberStyle = {
    ...gridItemStyle,
    backgroundColor: cellColors.roomNumber, // Change the color to your desired color
  };
  const pullMethodStyle = {
    ...gridItemStyle,
    backgroundColor: cellColors.pullMethod, // Change the color to your desired color
  };
  const handleCellClick = (roomNumber) => {
    setSelectedItem(roomNumber);
    setIsOpen(true);
  };
  const handlePullMethodChange = (e) => {
    setPullMethod(e.target.value);
  };
  const closeModal = () => {
    setShowModalError(false);
    setIsOpen(false);
  };
  const handleDropdownChange = (index, value) => {
    console.log(index);
    const updatedDropdownValues = [...dropdownValues];
    console.log(dropdownValues);
    updatedDropdownValues[index] = value;
    console.log(updatedDropdownValues);
    setDropdownValues(updatedDropdownValues);
  };

  const handleSubmit = (e) => {
    // Handle form submission logic here
    e.preventDefault();
    if (canIBump(pullMethod, selectedItem.notes)) {
        // valid bump
        const newRoomData = { roomNumber: selectedItem, notes: pullMethod, occupant1: dropdownValues[1], occupant2: dropdownValues[2], occupant3: dropdownValues[3] };
        console.log(newRoomData);
        updateGridData(newRoomData);
        console.log('Form submitted');
        setIsOpen(false);
    } else {
        // can't bump, show error 
        setShowModalError(true);
    }

    // // check that this is a valid pull method 


    
  };

//   const isUnbumpable = (notes) => {
//     if (notes == 'Preplaced') {
//       return true;
//     }
//     return false;
//   }

  const canIBump = (mine, yours) => {
    return true; // TODO

    // we can't bump preplaced, mentors, or proctors, or their pulls
    const forbiddenKeywords = ['preplaced', 'mentor', 'proctor']
    const containsForbiddenKeyword = (drawNumber) => {
        return forbiddenKeywords.some(keyword => drawNumber.toLowerCase().includes(keyword));
    };
    if (containsForbiddenKeyword(yours)) {
    return false;
    }

    // we can't bump people with in-dorm to that dorm or their pulls
    // take into account that the dorm to which they in dorm for matters

    // we can't bump people older than us
    const seniorityOrder = ['sophomore', 'junior', 'senior'];

  
    const getSeniorityIndex = (drawNumber) => {
      const seniorityMatch = drawNumber.match(/sophomore|junior|senior/i);
      return seniorityMatch ? seniorityOrder.indexOf(seniorityMatch[0].toLowerCase()) : -1;
    };
  
    const mySeniorityIndex = getSeniorityIndex(mine);
    const yourSeniorityIndex = getSeniorityIndex(yours);
  
    if (mySeniorityIndex === -1 || yourSeniorityIndex === -1) {
      return false; // Invalid draw numbers, don't have a date 
    }
  
    if (mySeniorityIndex === yourSeniorityIndex) {
      return Math.random() < 0.5; // Flip a coin for tie
    }
  
    return mySeniorityIndex > yourSeniorityIndex;
  };
    console.log(gridData);

  return (
    <div style={gridContainerStyle}>
        {/* <div style={gridItemStyle}><strong></strong></div>

        {/* begin filler code that does nothing*/}
        {/* <div style={gridItemStyle}><strong>{}</strong></div>
        <div style={gridItemStyle}><strong>{}</strong></div>
        <div style={gridItemStyle}><strong>{}</strong></div>
        <div style={gridItemStyle}><strong>{}</strong></div> */}




      <div style={roomNumberStyle}><strong>Room Number</strong></div>
      <div style={pullMethodStyle}><strong>Pull Method</strong></div>
      <div style={gridItemStyle}><strong>Occupant 1</strong></div>
      <div style={gridItemStyle}><strong>Occupant 2</strong></div>
      <div style={gridItemStyle}><strong>Occupant 3</strong></div>

      {gridData.map((item, index) => (
        <React.Fragment key={index}>
          <div style={roomNumberStyle} onClick={() => handleCellClick(item.roomNumber)}>{item.roomNumber}</div>
          <div style={pullMethodStyle} onClick={() => handleCellClick(item.roomNumber)}>{item.notes}</div>
          <div style={gridItemStyle} onClick={() => handleCellClick(item.roomNumber)}>{item.occupant1}</div>
          <div style={gridItemStyle} onClick={() => handleCellClick(item.roomNumber)}>{item.occupant2}</div>
          <div style={gridItemStyle} onClick={() => handleCellClick(item.roomNumber)}>{item.occupant3}</div>
        </React.Fragment>
      ))}
      {isOpen && (
        <div className="modal is-active">
          <div className="modal-background"></div>
          <div className="modal-card">
            <header className="modal-card-head">
              <p className="modal-card-title">Edit Room {selectedItem}</p>
              <button className="delete" aria-label="close" onClick={closeModal}></button>
            </header>
            <section className="modal-card-body">
                
            <div className="field">
                <label className="label" htmlFor="PullMethod">Pulling method: (ex: in dorm 11, in dorm 11 pull, sophomore 12)</label>
                <div className="control">
                  <input
                    id="PullMethod"
                    className="input"
                    type="text"
                    value={pullMethod}
                    onChange={handlePullMethodChange}
                  />
                </div>
              </div>
              {[1, 2, 3].map((index) => (
                <div className="field" key={index}>
                  <label className="label" htmlFor={`dropdown${index-1}`}>{`Occupant ${index}:`}</label>
                  <div className="control">
                    <div className="select">
                    <select value={dropdownValues[index]} onChange={(e) => handleDropdownChange(index, e.target.value)}>

                        <option value="">Select an option</option>
                        {dropdownOptions.map((option, optionIndex) => (
                          <option key={optionIndex} value={option}>{option}</option>
                        ))}
                      </select>
                    </div>
                  </div>
                </div>
              ))}
              
              {/* Add your modal content here */}
              {showModalError && (<p class="help is-danger">Oops-maybe double check your request?</p>)}

            </section>
            <footer className="modal-card-foot">
                
            <button className="button is-primary" onClick={handleSubmit}>Let's go!</button>

            </footer>
          </div>
        </div>
      )}
    </div>
    
  );
}

export default FloorGrid;
