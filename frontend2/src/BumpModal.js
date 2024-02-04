import React, { useState, useContext, useEffect } from 'react';
import { MyContext } from './MyContext';

function BumpModal() {
  const {
    selectedItem,
    // setSelectedItem,
    setIsModalOpen,
    selectedOccupants,
    setSelectedOccupants,
    pullMethod,
    setPullMethod,
    showModalError,
    setShowModalError,
    // drawNumbers,
    getNameById,
    userMap,
    setRefreshKey,
    refreshKey,
    selectedRoomObject,
    selectedSuiteObject,
    pullError,
    setPullError,
    // setSelectedRoomObject
  } = useContext(MyContext);

  const [othersInSuite, setOthersInSuite] = useState(["Person 1", "person 2"]);

  useEffect(() => {
    if (selectedSuiteObject) {
      const otherRooms = selectedSuiteObject.rooms;
      const otherOccupants = [];
      for (let room of otherRooms) {
        if (room.roomNumber !== selectedItem && room.maxOccupancy === 1 && room.occupant1 !== 0) {
          otherOccupants.push(room.occupant1);
        }
      }
      setOthersInSuite(otherOccupants);
    }
  }, [selectedSuiteObject, selectedItem]);      

  const handlePullMethodChange = (e) => {
    console.log(pullMethod);
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
    updatedselectedOccupants[index - 1] = value;
    //console.log(updatedselectedOccupants);
    setSelectedOccupants(updatedselectedOccupants);
    console.log(selectedOccupants);
  };

  const handleSubmit = async (e) => {  // Declare handleSubmit as async
    // Handle form submission logic here
    e.preventDefault();
    // if (!pullMethod) {
    //   setShowModalError(true);
    //   return false;
    // }
    if (await canIBump()) {  // Wait for canIBump to complete
      // TODO: check if you can bump
      // valid bump
      // const newRoomData = { roomNumber: selectedItem, notes: pullMethod, occupant1: selectedOccupants[1], occupant2: selectedOccupants[2], occupant3: selectedOccupants[3] };
      // //console.log(newRoomData);
      // updateGridData(newRoomData);
      //console.log('Form submitted');
      console.log("closed");
      closeModal();
    } else {
      console.log("showing error");
      // can't bump, show error 
      setShowModalError(true);
    }

    // // check that this is a valid pull method 
  };

  const canIBump = () => {
    return new Promise((resolve, reject) => {  // Return a new Promise
      fetch(`/rooms/${selectedRoomObject.roomUUID}`, {
        method: 'PATCH',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          proposedOccupants: selectedOccupants.map(Number).filter(num => num !== 0),  // Convert to array of numbers and remove zeros
          pullType: 1
        }),
      })
        .then(response => response.json())
        .then(data => {
          console.log(data);
          console.log("printed data");
          if (data.error) {
            console.log("error");
            console.log(data.error);
            setPullError(data.error);
            setIsModalOpen(true);
            setShowModalError(true);
            resolve(false);  // Resolve the Promise with false
          } else {
            console.log("empty ? room");
            console.log("setting refresh key");
            setRefreshKey(refreshKey + 1);
            resolve(true);  // Resolve the Promise with true
          }
        })
        .catch((error) => {
          console.log("hello");
          console.error(error.error);
          console.log("setting refresh key");
          setRefreshKey(refreshKey + 1);
          resolve(true);  // Resolve the Promise with false
        });

    });
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
          <label className="label">{`Occupant${selectedRoomObject.maxOccupancy > 1 ? "s" : ""}`}</label>



          {[1, 2, 3, 4].slice(0, selectedRoomObject.maxOccupancy).map((index) => (
            <div className="field" key={index}>
              <div className="control">
                <div className="select" style={{ marginRight: "10px" }}>
                  <select value={selectedOccupants[index - 1]} onChange={(e) => handleDropdownChange(index, e.target.value)}>

                    <option value="">Select an occupant</option>

                    {userMap && Object.keys(userMap)
                      .sort((a, b) => {
                        const nameA = `${userMap[a].FirstName} ${userMap[a].LastName}`;
                        const nameB = `${userMap[b].FirstName} ${userMap[b].LastName}`;
                        return nameA.localeCompare(nameB);
                      })
                      .map((key, index) => (
                        <option key={index} value={key}>
                          {userMap[key].FirstName} {userMap[key].LastName}
                        </option>
                      ))}

                  </select>
                </div>


              </div>

            </div>

          ))}
          <label className="label" >How did you pull this room?</label>
          <div className="select">
            <select value={pullMethod} onChange={handlePullMethodChange}>
              <option value="Select a pull method">Select a pull method</option>
              <option value="Pulled themselves">Pulled themselves</option>
              {othersInSuite.map((item, index) => (
                <option key={index} value={item}>
                  Pulled by {getNameById(item)}
                </option>
              ))}
              <option value="Lock Pull">Lock Pull</option>
            </select>
          </div>

          {/* Add your modal content here */}
          {showModalError && (<p class="help is-danger">{pullError}</p>)}

        </section>
        <footer className="modal-card-foot">

          <button className="button is-primary" onClick={handleSubmit}>Let's go!</button>

        </footer>
      </div>
    </div>
  );
}

export default BumpModal;