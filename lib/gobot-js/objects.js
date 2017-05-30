// specify the columns
var columnDefs = [
	  {headerName:'UUID', field:'UUID', width:80, editable:true, sortable:true}, 
	  {headerName:'Name', field:'Name', width:80, editable:true, sortable:true},
	  {headerName:'Bot Key', field:'BotKey', width:80, editable:true, sortable:true}, 
	  {headerName:'Bot Name', field:'BotName', width:80, editable:true, sortable:true},
	  {headerName:'Type', field:'Type', width:35, filter: 'number', editable:true},
	  {headerName:'Origin', field:'Origin', width:35, editable:true},
	  {headerName:'Position', field:'Position', width:60, editable:true, sortable:true},
	  {headerName:'Rotation', field:'Rotation', width:60, editable:true},
	  {headerName:'Velocity', field:'Velocity', width:60, editable:true},
	  {headerName:'Phantom', field:'Phantom', width:6, filter: 'number', editable:false},
	  {headerName:'Prims', field:'Prims', width:6, filter: 'number', editable:false},
	  {headerName:'BB High', field:'BBHigh', width:35, editable:false},
	  {headerName:'BB Low', field:'BBLow', width:35, editable:false},
	  {headerName:'LastUpdate', field:'LastUpdate', width:160, editable:true, sortable:true, sort:'desc'}
];

// get URLPathPrefix (hopefully it's already loaded by now)
var URLPathPrefix;

// let the grid know which columns and what data to use
var gridOptions = {
	columnDefs: columnDefs,
	rowData: null, // change to null since we're filling it from our own servers
	rowSelection: 'multiple',
	enableColResize: true,
	enableSorting: true,
	enableFilter: true,
	enableRangeSelection: true,
	suppressRowClickSelection: true,
	rowHeight: 22,
	animateRows: true,
	onModelUpdated: modelUpdated,
	debug: true,
	editType: 'fullRow',
    onRowValueChanged: function(event) {
        var data = event.data;
        console.log('onRowValueChanged: (' + data.UUID + ', ' + data.Name + ', ' + data.BotKey + ' ...)');
        // call the bloody fucking stuff from our REST API to update the database
        
        var httpRequest = new XMLHttpRequest(); // see https://stackoverflow.com/questions/6418220/javascript-send-json-object-with-ajax
		httpRequest.open('POST', URLPathPrefix + '/uiObjectsUpdate/');
		httpRequest.setRequestHeader("Content-Type", "application/json");
		httpRequest.onreadystatechange = function() { // see https://developer.mozilla.org/en-US/docs/Web/API/XMLHttpRequest/onreadystatechange
			if(httpRequest.readyState === XMLHttpRequest.DONE && httpRequest.status === 200) {
				console.log(httpRequest.responseText);
  			}
  		}
  		var response = JSON.stringify(data)
  		// console.log('Response is going to be: ' + response)
  		httpRequest.send(response);
    },
	onGridReady: function() {
		gridOptions.api.sizeColumnsToFit()
	}
};


// wait for the document to be loaded, otherwise ag-Grid will not find the div in the document.
document.addEventListener("DOMContentLoaded", function() {
	URLPathPrefix = document.getElementById("URLPathPrefix").innerText;

	// console.log('URLPathPrefix is "' + URLPathPrefix +'"');

	// lookup the container we want the Grid to use
	var eGridDiv = document.querySelector('#objectGrid');

	// create the grid passing in the div to use together with the columns & data we want to use
	new agGrid.Grid(eGridDiv, gridOptions);
	
	// do http request to get our sample data (see https://www.ag-grid.com/javascript-grid-value-getters/?framework=all#gsc.tab=0)
	var httpRequest = new XMLHttpRequest();
	httpRequest.open('GET', URLPathPrefix + '/uiObjects/');
	httpRequest.send();
	httpRequest.onreadystatechange = function() {
		if (httpRequest.readyState == 4 && httpRequest.status == 200) {
			var httpResult = JSON.parse(httpRequest.responseText);
			gridOptions.api.setRowData(httpResult);
		}
	};
});

function modelUpdated() {
}