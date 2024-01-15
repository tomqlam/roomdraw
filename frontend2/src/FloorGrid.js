import logo from './logo.svg';
import React from 'react';
import { useState, useContext } from 'react';
import { MyContext } from './MyContext';
import 'bulma/css/bulma.min.css';
import BumpModal from './BumpModal';

function FloorGrid({ cellColors, gridData, updateGridData }) {
  // Define your data structure with columns
  const { isModalOpen, setIsModalOpen } = useContext(MyContext);
  const { selectedItem, setSelectedItem } = useContext(MyContext);
  const { selectedOccupants, setSelectedOccupants } = useContext(MyContext);
  const { onlyShowBumpableRooms, setOnlyShowBumpableRooms } = useContext(MyContext);
  const { userMap } = useContext(MyContext);
  const { getNameById } = useContext(MyContext);
  const { selectedRoomObject, setSelectedRoomObject } = useContext(MyContext);
  const { pullMethod, setPullMethod } = useContext(MyContext);


  function getOccupantsByRoomNumber(roomNumber) {
    // Iterate over each suite
    for (let suite of gridData.suites) {
      // Find the room with the given room number within the current suite
      console.log("FIRST ITERATION");
      console.log(suite);
      const room = suite.rooms.find(r => r.roomNumber.toString() === roomNumber.toString());

      // If the room exists, return the list of occupants
      if (room) {
        setSelectedRoomObject(room);
        return [room.occupant1.toString(), room.occupant2.toString(), room.occupant3.toString(), room.occupant4.toString()];
      }
    }
    console.log("did not find the occupants");

    // If the room does not exist in any suite, return an empty array
    return ['', '', '', ''];
  }

  const gridContainerStyle = {
    display: 'grid',
    gridTemplateColumns: 'auto auto auto auto auto auto auto', // Adjust the number of 'auto' as needed
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
    fontSize: '15px', // Set the font size of the cells
    color: '#000000',

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
    setSelectedOccupants(getOccupantsByRoomNumber(roomNumber));
    console.log(selectedOccupants);
    console.log('lol');
    setPullMethod("");
    setIsModalOpen(true);
  };

  function getPullMethodByRoomNumber(roomNumber) {
    // TODO FINISH 
    // Iterate over each suite
    for (let suite of gridData.suites) {
      // Find the room with the given room number within the current suite
      const room = suite.rooms.find(r => r.roomNumber.toString() === roomNumber.toString());

      // If the room exists, return the list of occupants
      if (room) {
        if (room.pullPriority.isPreplaced) {
          return "Preplaced";
        }
        if (room.pullPriority.hasInDorm) {
          return `In-Dorm ${room.pullPriority.year} ${room.pullPriority.drawNumber}`;
        }
        return `${room.pullPriority.year} ${room.pullPriority.drawNumber}`;

      }
    }

    // If the room does not exist in any suite, return an empty array
    return 'n/a';
  }



  return (
    <div style={gridContainerStyle}>
      {/* <div style={gridItemStyle}><strong></strong></div>

        {/* begin filler code that does nothing*/}
      {/* <div style={gridItemStyle}><strong>{}</strong></div>
        <div style={gridItemStyle}><strong>{}</strong></div>
        <div style={gridItemStyle}><strong>{}</strong></div>
        <div style={gridItemStyle}><strong>{}</strong></div> */}




      <div style={roomNumberStyle}><strong>Room Number</strong></div>
      <div style={roomNumberStyle}><strong>Pull Method</strong></div>
      <div style={roomNumberStyle}><strong>Suite</strong></div>
      <div style={roomNumberStyle}><strong>Occupant 1</strong></div>
      <div style={roomNumberStyle}><strong>Occupant 2</strong></div>
      <div style={roomNumberStyle}><strong>Occupant 3</strong></div>
      <div style={roomNumberStyle}><strong>Occupant 4</strong></div>

      {gridData.suites.map((suite, suiteIndex) => (
        suite.rooms.map((room, roomIndex) => (
          <React.Fragment key={roomIndex}>
            <div
              style={{
                ...roomNumberStyle,
                backgroundColor: suiteIndex % 2 === 0
                  ? cellColors.evenSuite // color for even suiteIndex
                  : cellColors.oddSuite // color for odd suiteIndex
              }}
              onClick={() => handleCellClick(room.roomNumber)}
            >
              {room.roomNumber}
            </div>
            <div style={{
              ...pullMethodStyle, backgroundColor: suiteIndex % 2 === 0
                ? cellColors.evenSuite // color for even suiteIndex
                : cellColors.oddSuite
            }} onClick={() => handleCellClick(room.roomNumber)}>{getPullMethodByRoomNumber(room.roomNumber)}</div>
            {
              roomIndex === 0
              && <div style={{
                ...pullMethodStyle, gridRow: `span ${suite.rooms.length}`, backgroundColor: suiteIndex % 2 === 0
                  ? cellColors.evenSuite // color for even suiteIndex
                  : cellColors.oddSuite
              }} >Insert suite name</div>

            }
            <div style={{...gridItemStyle, backgroundColor: suiteIndex % 2 === 0 ? cellColors.evenSuite : cellColors.oddSuite}} onClick={() => handleCellClick(room.roomNumber)}>{getNameById(room.occupant1)}</div>
            <div style={{
              ...gridItemStyle,
              backgroundColor: room.maxOccupancy >= 2 ? (suiteIndex % 2 === 0 ? cellColors.evenSuite : cellColors.oddSuite) : cellColors.unbumpableRoom // Change the colors as per your requirement
            }} onClick={() => handleCellClick(room.roomNumber)}>{getNameById(room.occupant2)}</div>
            <div style={{
              ...gridItemStyle,
              backgroundColor: room.maxOccupancy >= 3 ? (suiteIndex % 2 === 0 ? cellColors.evenSuite : cellColors.oddSuite) : cellColors.unbumpableRoom // Change the colors as per your requirement
            }} onClick={() => handleCellClick(room.roomNumber)}>{getNameById(room.occupant3)}</div>
            {<div style={{
              ...gridItemStyle,
              backgroundColor: room.maxOccupancy >= 4 ? (suiteIndex % 2 === 0 ? cellColors.evenSuite : cellColors.oddSuite) : cellColors.unbumpableRoom // Change the colors as per your requirement
            }} onClick={() => handleCellClick(room.roomNumber)}>{getNameById(room.occupant4)}</div>}

          </React.Fragment>
        ))
      ))}





      {/* {gridData.map((item, index) => (
        <React.Fragment key={index}>
          <div style={roomNumberStyle} onClick={() => handleCellClick(item.roomNumber)}>{item.roomNumber}</div>
          <div style={pullMethodStyle} onClick={() => handleCellClick(item.roomNumber)}>{item.notes}</div>
          <div style={gridItemStyle} onClick={() => handleCellClick(item.roomNumber)}>{item.occupant1}</div>
          <div style={gridItemStyle} onClick={() => handleCellClick(item.roomNumber)}>{item.occupant2}</div>
          <div style={gridItemStyle} onClick={() => handleCellClick(item.roomNumber)}>{item.occupant3}</div>
        </React.Fragment>
      ))} */}
      {isModalOpen && <BumpModal updateGridData={updateGridData} />}
    </div>

  );
}

export default FloorGrid;
