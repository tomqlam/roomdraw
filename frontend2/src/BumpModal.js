import React, { useState, useContext, useEffect } from 'react';
import { MyContext } from './MyContext';

function BumpModal() {
  const {
    selectedItem,
    print,
    setIsModalOpen,
    selectedOccupants,
    setSelectedOccupants,
    pullMethod,
    setPullMethod,
    showModalError,
    setShowModalError,
    getNameById,
    userMap,
    setRefreshKey,
    refreshKey,
    selectedRoomObject,
    selectedSuiteObject,
    pullError,
    setPullError,
  } = useContext(MyContext);

  // List of arrays with two elements, where the first element is the occupant ID and the second element is the room UUID
  const [peopleWhoCanPull, setPeopleWhoCanPull] = useState([["Example ID", "Example Room UUID"]]);

  useEffect(() => {
    // If the selected suite or room changes, change the people who can pull 
    if (selectedSuiteObject) {
      const otherRooms = selectedSuiteObject.rooms;
      const otherOccupants = [];
      for (let room of otherRooms) {
        if (room.roomNumber !== selectedItem && room.maxOccupancy === 1 && room.occupant1 !== 0 && room.pullPriority.pullType === 1) {
          otherOccupants.push([room.occupant1, room.roomUUID]);
        }
      }
      setPeopleWhoCanPull(otherOccupants);
    }
  }, [selectedSuiteObject, selectedItem]);      

  const handlePullMethodChange = (e) => {
    print(pullMethod);
    setPullMethod(e.target.value);
  };
  const closeModal = () => {
    setShowModalError(false);
    setIsModalOpen(false);
  };
  const handleDropdownChange = (index, value) => {
    
    const updatedselectedOccupants = [...selectedOccupants];
    updatedselectedOccupants[index - 1] = value;
    setSelectedOccupants(updatedselectedOccupants);

  };

  const handleSubmit = async (e) => {  // Declare handleSubmit as async
    // Handle form submission logic here
    e.preventDefault();
    
    if (/^\d+$/.test(pullMethod)) {
      console.log("Pull method is a number");
      // pullMethod only includes number, implying that you were pulled by someone else
      if (await canIBePulled()) {  // Wait for canIBePulled to complete
        print("This room was successfully pulled by someone else in the suite");
        closeModal();
      } else {
        setShowModalError(true);
      }
    } else {
      // pullMethod is either Lock Pull or Pulled themselves 
      if (await canIBump()) {  // Wait for canIBePulled to complete

        print("This room pulled themselves");
        closeModal();
      } else {
        setShowModalError(true);
      }

    }
    

  };

  const canIBump = () => {
    // for non-normal pulls, this checks if the room can be pulled
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
          if (data.error) {
            // There was an error pulling the room
            setPullError(data.error);
            setIsModalOpen(true);
            setShowModalError(true);
            resolve(false); 
          } else {
            // There was no error pulling the room, and the grid should refresh
            setRefreshKey(refreshKey + 1);
            resolve(true); 
          }
        })
        .catch((error) => {
          print(error.error);
          setRefreshKey(refreshKey + 1);
          resolve(true); 
        });

    });
  }

  const canIBePulled = () => {
    return new Promise((resolve, _) => {  // Return a new Promise
      fetch(`/rooms/${selectedRoomObject.roomUUID}`, {
        method: 'PATCH',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          proposedOccupants: selectedOccupants.map(Number).filter(num => num !== 0),  // Convert to array of numbers and remove zeros
          pullType: 2,
          pullLeaderRoom: peopleWhoCanPull.find(person => person[0] === Number(pullMethod))[1],
        }),
      })
        .then(response => response.json())
        .then(data => {
          if (data.error) {
            // There was an error pulling the room
            setPullError(data.error);
            setIsModalOpen(true);
            setShowModalError(true);
            resolve(false); 
          } else {
            // There was no error pulling the room, and the grid should refresh
            setRefreshKey(refreshKey + 1);
            resolve(true); 
          }
        })
        .catch((error) => {
          console.error(error.error);
          setRefreshKey(refreshKey + 1);
          resolve(true); 
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

                    <option value="">No occupant</option>

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
              <option value="Pulled themselves">Pulled themselves</option>
              {selectedRoomObject.maxOccupancy === 1 && peopleWhoCanPull.map((item, index) => (
                <option key={index} value={item[0]}>
                  Pulled by {getNameById(item[0])}
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