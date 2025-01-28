<?php

namespace App\Http\Controllers;

use Illuminate\Http\Request;
use App\Models\Log;

class LogController extends Controller
{
    




    public function store(Request $request){
           

        $log = new Log();
        $log -> agent = $request('agent');
        $log -> Event = $request('Event');
        $log -> Details = $request('Details');
        $log -> save();


        return response() -> json(['message' => 'LOG created'],200);


    }







}
