import React, { useState, useContext } from 'react';
import { MyContext } from './MyContext';

function BumpModal({ updateGridData }) {
  const { selectedItem, setSelectedItem } = useContext(MyContext);
  const { setIsModalOpen } = useContext(MyContext);
  const { selectedOccupants, setSelectedOccupants } = useContext(MyContext);
  const { pullMethod, setPullMethod } = useContext(MyContext);
  const { showModalError, setShowModalError } = useContext(MyContext);
  const { drawNumbers } = useContext(MyContext);
  const { getNameById } = useContext(MyContext);
  const { userMap } = useContext(MyContext);
  const {selectedRoomObject, setSelectedRoomObject} = useContext(MyContext);
  const handlePullMethodChange = (e) => {
    setPullMethod(e.target.value);
  };
  const closeModal = () => {
    setShowModalError(false);
    setIsModalOpen(false);
  };
  const handleDropdownChange = (index, value) => {
    console.log(value);
    const updatedselectedOccupants = [...selectedOccupants];
    console.log(selectedOccupants);
    //console.log(selectedOccupants);
    updatedselectedOccupants[index-1] = value;
    //console.log(updatedselectedOccupants);
    setSelectedOccupants(updatedselectedOccupants);
  };

  const handleSubmit = (e) => {
    // Handle form submission logic here
    e.preventDefault();
    if (!pullMethod) {
      setShowModalError(true);
      return false;
    }
    console.log("got here");
    if (canIBump()) {
      console.log("allowd to bunp");
      // TODO: check if you can bump
      // valid bump
      // const newRoomData = { roomNumber: selectedItem, notes: pullMethod, occupant1: selectedOccupants[1], occupant2: selectedOccupants[2], occupant3: selectedOccupants[3] };
      // //console.log(newRoomData);
      // updateGridData(newRoomData);
      //console.log('Form submitted');
      setIsModalOpen(false);
    } else {
      // can't bump, show error 
      setShowModalError(true);
    }

    // // check that this is a valid pull method 
  };

  const canIBump = () => {
    // todo
    return true;
  }


  return (
    <div className="modal is-active">
      <div className="modal-background"></div>
      <div className="modal-card">
        <header className="modal-card-head">
          <p className="modal-card-title">Edit Room {selectedItem}</p>
          <button className="delete" aria-label="close" onClick={closeModal}></button>
        </header>
        <section className="modal-card-body">
          {/* description */}


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
          {[1, 2, 3, 4].slice(0,selectedRoomObject.maxOccupancy).map((index) => (
            <div className="field" key={index}>
              <label className="label" htmlFor={`dropdown${index - 1}`}>{`Occupant ${index}:`}</label>
              <div className="control">
                <div className="select">
                  <select value={selectedOccupants[index-1]} onChange={(e) => handleDropdownChange(index, e.target.value)}>

                    <option value="">Select an option</option>
                    {/* {users.map((user) => `${user.FirstName}`).map((option, optionIndex) => (
                      <option key={optionIndex} value={option}>{option}</option>
                    ))} */}
                        {Object.keys(userMap).map((key, index) => (
                    <option key={index} value={key}>
                      {userMap[key].FirstName} {userMap[key].LastName} 
                    </option>
                  ))}
                    {/* {drawNumbers.map((number) => number.name).map((option, optionIndex) => (
                      <option key={optionIndex} value={option}>{option}</option>
                    ))} */}
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
  );
}

export default BumpModal;