// specify the columns
var columnDefs = [
	{headerName:'UUID', field:'UUID', width:80, editable:true, sortable:true}, 
	{headerName:'Name', field:'Name', width:80, editable:true, sortable:true},
	{headerName:'OwnerKey', field:'UUID', width:80, editable:true, sortable:true}, 
	{headerName:'OwnerName', field:'Name', width:80, editable:true, sortable:true},
	{headerName:'Location', field:'Location', width:60, editable:true, sortable:true},
	{headerName:'Position', field:'Position', width:60, editable:true, sortable:true},
	{headerName:'Rotation', field:'Rotation', width:60, editable:true},
	{headerName:'Velocity', field:'Velocity', width:60, editable:true},
	{headerName:'Energy', field:'Energy', width:35, editable:true, sortable:true}, 
	{headerName:'Money', field:'Money', width:35, editable:true, sortable:true},
	{headerName:'Happiness', field:'Happiness', width:35, editable:true, sortable:true},
	{headerName:'Class', field:'Class', width:35, editable:true, sortable:true},
	{headerName:'SubType', field:'SubType', width:35, editable:true},
	{headerName:'PermURL', field:'PermURL', width:80, editable:true},
	{headerName:'LastUpdate', field:'LastUpdate', width:160, editable:true, sortable:true},
	{headerName:'BestPath', field:'BestPath', width:30, editable:true, sortable:true},
	{headerName:'SecondBestPath', field:'SecondBestPath', width:30, editable:true, sortable:true},
	{headerName:'CurrentTarget', field:'CurrentTarget', width:60, editable:true, sortable:true},		 
];


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
	onGridReady: function() {
		gridOptions.api.sizeColumnsToFit()
	}
};


// wait for the document to be loaded, otherwise ag-Grid will not find the div in the document.
document.addEventListener("DOMContentLoaded", function() {

	// lookup the container we want the Grid to use
	var eGridDiv = document.querySelector('#agentGrid');

	// create the grid passing in the div to use together with the columns & data we want to use
	new agGrid.Grid(eGridDiv, gridOptions);
	
	// do http request to get our sample data (see https://www.ag-grid.com/javascript-grid-value-getters/?framework=all#gsc.tab=0)
	var httpRequest = new XMLHttpRequest();
	httpRequest.open('GET', '/go/uiAgents/'); // should have /go/ as variable somewhere... (20170514)
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